package tables

import (
	"errors"
	"log"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/wazupwiddat/poker-server/server/cards"
)

type Round int

const (
	PreFlop Round = iota
	Flop
	Turn
	River
)

type SaveHand struct {
	GameKey        string
	ID             int
	Winners        []string
	WinDescription string
	Cards          []string
	Community      []string
}

type TexasHoldemCardGame struct {
	handCount   int
	updates     func(GameUpdate)
	actionTimer *time.Timer
	active      *Player
	table       *Table
	deck        *cards.Deck
	cards       []cards.Card
	round       Round
	button      int
	cost        int
	lastWinner  *cards.Hand
	winners     []*Player
}

func (c *TexasHoldemCardGame) Updates(f func(GameUpdate)) {
	if f != nil {
		c.updates = f
	}
}

func (c *TexasHoldemCardGame) ForceUpdate() {
	c.doUpdate()
}

func (c *TexasHoldemCardGame) doUpdate() {
	if c.updates != nil {
		c.updates(GameUpdate{
			Active:     c.active,
			Button:     c.button,
			Cards:      c.cards,
			Cost:       c.cost,
			Round:      c.round,
			Pots:       c.pots(),
			LastWinner: c.lastWinner,
			Winners:    c.winners,
		})
	}
}

func (c *TexasHoldemCardGame) StartRound() {
	log.Println("Setup Round")
	if c.round == PreFlop {
		c.sitoutPlayers() // force players with chips to sitout
	}

	if c.occupiedSeats() < 2 {
		if c.actionTimer != nil {
			c.actionTimer.Stop()
		}
		c.doUpdate()
		log.Println("not enough to play")
		return
	}
	c.resetAction()

	switch c.round {
	case PreFlop:
		c.button = c.nextSeat(c.button)
		sb := c.nextSeat(c.button)
		bb := c.nextSeat(sb)
		if c.occupiedSeats() == 2 {
			bb = c.button
			sb = c.nextSeat(c.button)
		}
		// brand new suffled deck
		c.deck = c.table.dealer.Deck()
		c.dealCards()
		c.table.Seats[sb].contribute(c.table.Options.Stakes.SmallBlind)
		c.table.Seats[bb].contribute(c.table.Options.Stakes.BigBlind)
		firstAction := c.nextSeat(bb)
		c.active = c.table.Seats[firstAction]
		c.cost = c.table.Options.Stakes.BigBlind
		c.cards = nil
	case Flop:
		c.cards = c.deck.PopMulti(3)
		c.active = c.table.Seats[c.button]
		seat := c.nextToAct()
		c.active = c.table.Seats[seat]
		c.cost = 0
	case Turn, River:
		c.cards = append(c.cards, c.deck.Pop())
		c.active = c.table.Seats[c.button]
		seat := c.nextToAct()
		c.active = c.table.Seats[seat]
		c.cost = 0
	}
	c.resetActionTimer()
	c.doUpdate()
}

func (c *TexasHoldemCardGame) sitoutPlayers() {

	newbt := c.nextSeat(c.button)
	sb := c.nextSeat(newbt)
	bb := c.nextPossibleSeat(sb) // am I now allowed to come in

	for i, seat := range c.table.Seats {
		if seat == nil {
			continue
		}
		if seat.ChipCount == 0 {
			seat.SittingOut = true
			continue
		}
		if seat.SittingOut && !seat.SittingOutNextHand && i == bb {
			log.Println("Seat missed BB", i)
			seat.SittingOut = false
			continue
		}
		if seat.SittingOutNextHand {
			seat.SittingOut = true
		}
	}
}

func (c *TexasHoldemCardGame) dealCards() {
	for _, seat := range c.table.Seats {
		if seat == nil {
			continue
		}
		seat.ChipsInPot = 0
		seat.ChipsInRound = 0
		seat.Acted = false
		seat.Folded = false
		seat.AllIn = false
		seat.Reveal = nil
		seat.Cards = nil
		if !seat.SittingOut {
			seat.Cards = c.deck.PopMulti(2)
			seat.contribute(c.table.Options.Stakes.Ante)
		}
	}
}

func (c *TexasHoldemCardGame) Act(player *Player, act Action) error {
	if player != c.active || player.Acted {
		return errors.New("Game.Act: it is not your turn")
	}
	if c.occupiedSeats() < 2 {
		if c.actionTimer != nil {
			c.actionTimer.Stop()
		}
		c.doUpdate()
		log.Println("not enough to play")
		return nil
	}

	log.Println("Table.Act", act)
	if includes(c.legalActions(), act.Type) == false {
		return errors.New("Game.Act: illegal action attempted")
	}
	switch act.Type {
	case Fold:
		c.active.Folded = true
	case Check:
	case Call:
		c.active.contribute(c.owed())
	case Bet, Raise:
		if c.cost == 0 && act.Chips < c.table.Options.Stakes.BigBlind {
			return errors.New("Game.Act: bet needs to be at least big blind")
		}
		c.active.contribute(c.owed())
		c.active.contribute(act.Chips)
		c.resetAction()
	case AllIn:
		c.active.contribute(c.owed())
		c.active.contribute(c.active.ChipCount)
		c.resetAction()
	}
	c.active.Acted = true
	c.active.Action = act.Type
	if c.active.ChipsInRound > c.cost {
		c.cost = c.active.ChipsInRound
	}

	// go routine is because of our stupid sleep where players have time to show
	// and the player who last folded socket is in use so his reveal cards would happen after
	// the next round started :(
	go c.roundUpdate()
	return nil
}

func (c *TexasHoldemCardGame) Stop() {
	if c.actionTimer != nil {
		c.actionTimer.Stop()
	}
}

func (c *TexasHoldemCardGame) resetActionTimer() {
	if c.actionTimer != nil {
		c.actionTimer.Stop()
	}
	c.actionTimer = time.NewTimer(time.Second * 30)
	go func() {
		<-c.actionTimer.C
		if includes(c.legalActions(), Check) {
			c.Act(c.active, Action{
				Type: Check,
			})
			return
		}
		c.Act(c.active, Action{
			Type: Fold,
		})
	}()
}

func (c *TexasHoldemCardGame) resetAction() {
	for _, seat := range c.table.Seats {
		if seat != nil {
			seat.Acted = false
			seat.Action = -1
		}
	}
}

func (c *TexasHoldemCardGame) resetChipsInRound() {
	for _, seat := range c.table.Seats {
		if seat != nil {
			seat.ChipsInRound = 0
		}
	}
}

func (c *TexasHoldemCardGame) legalActions() []ActionType {
	if c.cost == 0 {
		return []ActionType{Fold, Check, Bet, AllIn}
	}
	if c.owed() == 0 {
		return []ActionType{Fold, Check, Bet, AllIn}
	}
	if c.owed() > c.active.ChipCount {
		return []ActionType{Fold, Call}
	}
	return []ActionType{Fold, Call, Raise, AllIn}
}

func (c *TexasHoldemCardGame) owed() int {
	return c.cost - c.active.ChipsInRound
}

func (c *TexasHoldemCardGame) roundUpdate() {
	log.Println("roundUpdate")
	// everyone folded scenario
	if len(c.contesting()) == 1 {
		c.payout()
		c.round = PreFlop
		c.StartRound()
		return
	}
	// contest continues, get next to act
	seat := c.nextToAct()
	if seat != -1 {
		c.active = c.table.Seats[seat]
		c.resetActionTimer()
		c.doUpdate()
		log.Println("roundUpdate active seat is ", seat)
		return
	}
	log.Println("round update all have acted")
	c.resetChipsInRound()
	// all have acted for this round
	if c.round == River {
		log.Println("Showing cards for winner")
		c.revealContesting()
		c.payout()
		c.round = PreFlop
	} else {
		c.round = (c.round + 1) % (River + 1)
	}
	c.StartRound()
}

func (c *TexasHoldemCardGame) payout() {
	log.Println("payout")
	c.handCount++
	if len(c.contesting()) == 1 {
		pots := c.pots()
		if len(pots) == 1 {
			c.payoutToWinners(c.contesting(), pots[0])
			c.saveHands()
			c.lastWinner = nil
			c.winners = nil
			c.doUpdate()
			return
		}
		log.Println("Payout Error", "pots should be 1 when only 1 contests")
		return
	}
	hands := map[*Player]*cards.Hand{}
	for _, seat := range c.table.Seats {
		if seat == nil {
			continue
		}
		hands[seat] = cards.NewHand(append(seat.Cards, c.cards...))
	}
	for _, pot := range c.pots() {
		// sort by best hand first
		sort.Slice(pot.Contesting, func(i, j int) bool {
			iHand := hands[pot.Contesting[i]]
			jHand := hands[pot.Contesting[j]]
			return iHand.CompareTo(jHand) > 0
		})
		// select winners who split pot if more than one
		winners := []*Player{}
		h1 := hands[pot.Contesting[0]]
		c.lastWinner = h1
		for _, seat := range pot.Contesting {
			h2 := hands[seat]
			if h1.CompareTo(h2) != 0 {
				break
			}
			winners = append(winners, seat)
		}
		// sort closest to the button for spare chips in split pot
		sort.Slice(winners, func(i, j int) bool {
			iDist := c.distanceFromButton(winners[i])
			jDist := c.distanceFromButton(winners[j])
			return iDist < jDist
		})
		// payout chips
		c.payoutToWinners(winners, pot)
	}
	c.saveHands()
	c.lastWinner = nil
	c.winners = nil
	c.doUpdate()
}

func (c *TexasHoldemCardGame) payoutToWinners(winners []*Player, pot *SidePot) {
	for i, seat := range winners {
		seat.ChipCount += pot.Chips / len(winners)
		if (pot.Chips % len(winners)) > i {
			seat.ChipCount++
		}
	}
	c.winners = winners
	c.doUpdate()
	time.Sleep(time.Second * 5) // cheazy
}

func (c *TexasHoldemCardGame) saveHands() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	var winners []string
	var cards []string
	if len(c.winners) > 0 {
		for _, p := range c.winners {
			winners = append(winners, p.Name)
			if len(cards) == 0 {
				for _, pc := range p.Reveal {
					cards = append(cards, pc.String())
				}
			}
		}
	}

	var community []string
	if len(c.cards) > 0 {
		for _, ca := range c.cards {
			community = append(community, ca.String())
		}
	}
	desc := ""
	if c.lastWinner != nil {
		desc = c.lastWinner.Description()
	}

	saveItem := SaveHand{
		GameKey:        c.table.ID,
		ID:             c.handCount,
		Winners:        winners,
		WinDescription: desc,
		Cards:          cards,
		Community:      community,
	}

	av, err := dynamodbattribute.MarshalMap(saveItem)
	if err != nil {
		log.Println("Got error marshalling map:")
		log.Println(err.Error())
		return
	}
	// Create item in table Movies
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String("hands"),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		log.Println("Got error calling PutItem:")
		log.Println(err.Error())
	}
	log.Println("Hand Saved", c.table.ID, c.handCount)
}

func (c *TexasHoldemCardGame) pots() []*SidePot {
	contesting := c.contesting()
	sort.Slice(contesting, func(i, j int) bool {
		return contesting[i].ChipsInPot < contesting[j].ChipsInPot
	})
	costs := []int{}
	for _, seat := range contesting {
		if contains(costs, seat.ChipsInPot) == false {
			costs = append(costs, seat.ChipsInPot)
		}
	}
	pots := []*SidePot{}
	for i, cost := range costs {
		pot := &SidePot{}
		min := 0
		if i != 0 {
			min = costs[i-1]
		}
		for _, seat := range c.table.Seats {
			if seat == nil {
				continue
			}
			pot.Chips += max(seat.ChipsInPot-min, 0)
		}
		for _, seat := range contesting {
			if seat.ChipsInPot >= cost {
				pot.Contesting = append(pot.Contesting, seat)
			}
		}
		pots = append(pots, pot)
	}
	return pots
}

func (c *TexasHoldemCardGame) contesting() []*Player {
	contesting := []*Player{}
	for _, seat := range c.table.Seats {
		if seat != nil && seat.Folded == false && !seat.SittingOut {
			contesting = append(contesting, seat)
		}
	}
	return contesting
}

func (c *TexasHoldemCardGame) revealContesting() {
	for _, seat := range c.table.Seats {
		if seat != nil && seat.Folded == false {
			seat.Reveal = seat.Cards
		}
	}
	c.doUpdate()
}

func (c *TexasHoldemCardGame) distanceFromButton(p *Player) int {
	seat := c.button
	dist := 0
	for i := 0; i < len(c.table.Seats); i++ {
		seat = c.nextSeat(seat)
		dist++
		if p.Seat == seat {
			return dist
		}
	}
	return -1
}

func (c *TexasHoldemCardGame) nextToAct() int {
	seat := c.active.Seat
	for i := 0; i < len(c.table.Seats); i++ {
		seat = c.nextSeat(seat)
		p := c.table.Seats[seat]
		if !p.Acted && !p.AllIn && !p.Folded {
			return p.Seat
		}
	}
	return -1
}

func (c *TexasHoldemCardGame) nextPossibleSeat(seat int) int {
	for i := 0; i < len(c.table.Seats); i++ {
		seat = (seat + 1) % len(c.table.Seats)
		p := c.table.Seats[seat]
		if p != nil {
			return seat
		}
	}
	return -1
}

func (c *TexasHoldemCardGame) nextSeat(seat int) int {
	for i := 0; i < len(c.table.Seats); i++ {
		seat = (seat + 1) % len(c.table.Seats)
		p := c.table.Seats[seat]
		if p != nil && !p.SittingOut {
			return seat
		}
	}
	return -1
}

func (c *TexasHoldemCardGame) occupiedSeats() int {
	cnt := 0
	for _, seat := range c.table.Seats {
		if seat != nil && !seat.SittingOut {
			cnt++
		}
	}
	return cnt
}

func contains(a []int, i int) bool {
	for _, v := range a {
		if v == i {
			return true
		}
	}
	return false
}

func max(i, j int) int {
	if i > j {
		return i
	}
	return j
}

func includes(actions []ActionType, include ...ActionType) bool {
	for _, a1 := range include {
		found := false
		for _, a2 := range actions {
			found = found || a1 == a2
		}
		if !found {
			return false
		}
	}
	return true
}
