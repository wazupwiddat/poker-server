package tables

import (
	"errors"
	"log"
	"math/rand"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	socketio "github.com/googollee/go-socket.io"

	"github.com/wazupwiddat/poker-server/server/cards"
)

type Variant int

const (
	TexasHoldem Variant = iota
	TexasHoldem2
	OmahaHi
)

type Limit int

const (
	NoLimit Limit = iota
	PotLimit
)

type Options struct {
	Buyin   int
	Variant Variant
	Stakes  Stakes
	Limit   Limit
}

var (
	DefaultOptions = Options{
		Buyin:   200,
		Variant: TexasHoldem,
		Stakes: Stakes{
			BigBlind:   2,
			SmallBlind: 1,
			Ante:       0,
		},
		Limit: NoLimit,
	}
)

type Stakes struct {
	BigBlind   int
	SmallBlind int
	Ante       int
}

type Chat struct {
	User    string
	Message string
}

type TableUpdate struct {
	Seats      []*Player
	GameUpdate GameUpdate
}

type Table struct {
	ID      string
	Options Options

	chat        chan Chat
	tableUpdate chan TableUpdate
	save        chan SaveTableItem

	sockets map[string]socketio.Conn
	tick    *time.Ticker

	OnClose func(t *Table)

	// poker stuff
	game      CardGame
	gameRound Round

	dealer cards.Dealer
	Seats  []*Player
}

type SavePlayer struct {
	Name      string
	Seat      int
	BuyIn     int
	ChipCount int
}

type SaveTableItem struct {
	GameKey string
	Players []*SavePlayer
}

func NewTable(id string) *Table {
	s := make(map[string]socketio.Conn)
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)

	t := &Table{
		ID:          id,
		tableUpdate: make(chan TableUpdate),
		chat:        make(chan Chat),
		save:        make(chan SaveTableItem),
		tick:        time.NewTicker(time.Second * 10),
		sockets:     s,
		Seats:       make([]*Player, 9),
		Options:     DefaultOptions,
		dealer:      cards.NewDealer(r),
	}
	return t
}

func (t *Table) UpdateAll() {
	if t.game == nil {
		// before game exists update only seats
		t.tableUpdate <- TableUpdate{
			Seats: t.Seats,
		}
		return
	}
	t.game.ForceUpdate()
}

func (t *Table) triggerTableSave() {
	savePlayers := []*SavePlayer{}
	for _, p := range t.Seats {
		if p == nil || p.Name == "" {
			continue
		}
		sp := &SavePlayer{
			Seat:      p.Seat,
			Name:      p.Name,
			BuyIn:     p.BuyIn,
			ChipCount: p.ChipCount,
		}
		savePlayers = append(savePlayers, sp)
	}
	t.save <- SaveTableItem{
		GameKey: t.ID,
		Players: savePlayers,
	}
}

func (t *Table) StartDealing() {
	t.game = NewCardGame(t)

	// connect game updates to the tableUpdate channel
	t.game.Updates(func(update GameUpdate) {
		t.tableUpdate <- TableUpdate{
			Seats:      t.Seats,
			GameUpdate: update,
		}
		if t.gameRound != update.Round && update.Round == PreFlop {
			t.triggerTableSave()
		}
		t.gameRound = update.Round
	})
	// start the first round
	t.game.StartRound()
}

func (t *Table) Act(sid string, act Action) error {
	if t.game == nil {
		return errors.New("Table.Act: Game has not been setup")
	}

	p := t.findPlayer(sid)
	if p == nil {
		return errors.New("Table.Act: player does not exist for action request: " + sid)
	}
	return t.game.Act(p, act)

}

func (t *Table) AddPlayerSocket(s socketio.Conn) {
	if s.ID() == "" {
		return
	}
	t.sockets[s.ID()] = s
}

func (t *Table) AddPlayerToSeat(sid string, seat int) bool {
	if t.Seats[seat] != nil {
		return false
	}
	t.Seats[seat] = NewPlayer(sid, seat)
	if t.game != nil { // game already started, sit out player
		t.Seats[seat].SittingOut = true
		t.Seats[seat].SittingOutNextHand = false
	}
	t.UpdateAll()
	t.triggerTableSave()
	return true
}

func (t *Table) RemovePlayerFromSeat(sid string) {
	for k, v := range t.Seats {
		if v == nil {
			continue
		}
		if v.ID == sid {
			t.Seats[k] = nil
			t.UpdateAll()
			t.triggerTableSave()
			return
		}
	}
}

func (t *Table) RemovePlayerSocket(sID string) {
	delete(t.sockets, sID)
	// if last socket close everything down and turn out the lights
	if len(t.sockets) == 0 {
		close(t.tableUpdate)
		close(t.chat)
		close(t.save)
		t.tick.Stop()
		if t.game != nil {
			t.game.Stop()
			t.game = nil
		}

		// inform PokerServer (Owner) we are done here
		if t.OnClose != nil {
			t.OnClose(t)
		}
	}
}

func (t *Table) findPlayer(id string) *Player {
	for _, p := range t.Seats {
		if p == nil {
			continue
		}
		if p.ID == id {
			return p
		}
	}
	return nil
}

func (t *Table) modifyPlayer(id string, f func(*Player)) {
	for _, p := range t.Seats {
		if p == nil {
			continue
		}
		if p.ID == id {
			f(p)
			t.UpdateAll()
			t.triggerTableSave()
			return
		}
	}
}

func (t *Table) PeekAtCards(sid string) []cards.Card {
	log.Println("peeking at cards", sid)
	p := t.findPlayer(sid)
	if p == nil {
		log.Println("peekAtCards error", sid)
		return nil
	}
	return p.Cards
}

func (t *Table) RevealMyCards(sid string) {
	log.Println("revealing my cards", sid)
	t.modifyPlayer(sid, func(p *Player) {
		p.Reveal = p.Cards
	})
}

func (t *Table) UpdateUsernameToPlayer(sid string, uname string) {
	log.Println("modifying the player username")
	t.modifyPlayer(sid, func(p *Player) {
		p.Name = uname
	})
}

func (t *Table) AddChipsToPlayer(sid string, amt int) {
	log.Println("modifying the players buy-in and chip count", amt)
	t.modifyPlayer(sid, func(p *Player) {
		p.BuyIn += amt
		p.ChipCount += amt
	})
}

func (t *Table) HandleTableUpdates() {
	log.Println("Table update handler waiting...")
	for {
		tableUpdate, ok := <-t.tableUpdate
		if !ok {
			log.Println("Table update hanlers closed")
			break
		}
		for _, v := range t.sockets {
			v.Emit("tableUpdate", tableUpdate)
		}
	}
}

func (t *Table) HandleChatMessages() {
	// as state changes or action occurs
	log.Println("Table handle action chat...")
	for {
		chat, ok := <-t.chat
		if !ok {
			log.Println("Chat channel closed", t.ID)
			break
		}
		for _, v := range t.sockets {
			v.Emit("chatResponse", chat)
		}
	}
	log.Println("Table handle chat closed")
}

func (t *Table) HandleSavingTableData() {
	log.Println("Table handle saving table data...")
	for {
		save, ok := <-t.save
		if !ok {
			log.Println("Save channel closed", t.ID)
			break
		}
		t.SaveTable(save)
	}
	log.Println("Table handle save table data closed")
}

func (t *Table) PrintInfo() {
	log.Println("Table info:")
	log.Println("  ID = ", t.ID)
	log.Println("  Connected = ", len(t.sockets))
	scnt := 0
	for _, s := range t.Seats {
		if s != nil {
			scnt++
		}
	}
	log.Println("  Seated = ", scnt)
}

func (t *Table) Sitout(sid string) {
	log.Println("Player sitting out", sid)
	p := t.findPlayer(sid)
	if p == nil {
		return
	}
	p.SittingOutNextHand = true
	t.UpdateAll()
}

func (t *Table) Sitin(sid string) {
	log.Println("Player sitting in", sid)
	p := t.findPlayer(sid)
	if p == nil {
		return
	}
	p.SittingOutNextHand = false
	t.UpdateAll()
}

func (t *Table) Ping() {
	log.Println("Table ping...")
	for range t.tick.C {
		t.PrintInfo()
	}
	log.Println("Table ping closed")
}

func (t *Table) SendChat(sid, msg string) {
	p := t.findPlayer(sid)
	if p == nil {
		return
	}
	t.chat <- Chat{User: p.Name, Message: msg}
}

func (t *Table) SaveTable(saveItem SaveTableItem) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	getItemOutput, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("tables"),
		Key: map[string]*dynamodb.AttributeValue{
			"GameKey": {
				S: aws.String(t.ID),
			},
		},
	})
	if err != nil {
		log.Println("Table.SaveTable:", err)
		return
	}
	updateItem := SaveTableItem{}
	err = dynamodbattribute.UnmarshalMap(getItemOutput.Item, &updateItem)
	if err != nil {
		log.Println("Table.SaveTable:", err)
		return
	}

	var av map[string]*dynamodb.AttributeValue
	if updateItem.GameKey == saveItem.GameKey {
		if ok := mergeUpdateSaveTableItem(&updateItem, saveItem); !ok {
			log.Println("Table.SaveTable:", "had a problem merging the table data")
			return
		}
		av, err = dynamodbattribute.MarshalMap(updateItem)
		if err != nil {
			log.Println("Got error marshalling map:")
			log.Println(err.Error())
			return
		}
	} else {
		av, err = dynamodbattribute.MarshalMap(saveItem)
	}
	// Create item in table "tables"
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String("tables"),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		log.Println("Got error calling PutItem:")
		log.Println(err.Error())
	}
	log.Println("Table Saved", t.ID)
}

func mergeUpdateSaveTableItem(src *SaveTableItem, new SaveTableItem) bool {
	if src.GameKey != new.GameKey {
		log.Println("mergeUpdateSaveTableItem source and update items are not for the same table")
		return false
	}

	// new vs update
	for _, np := range new.Players {
		if np == nil {
			continue
		}
		sp := findSavePlayer(src.Players, np.Name, np.Seat)
		if sp == nil { // new append
			src.Players = append(src.Players, np)
			continue
		}
		sp.BuyIn = np.BuyIn // exists just update
		sp.ChipCount = np.ChipCount
	}
	return true
}

func findSavePlayer(src []*SavePlayer, name string, seat int) *SavePlayer {
	for _, sp := range src {
		if sp.Name == name && sp.Seat == seat {
			return sp
		}
	}
	return nil
}
