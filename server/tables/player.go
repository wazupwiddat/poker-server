package tables

import "github.com/wazupwiddat/poker-server/server/cards"

type Player struct {
	ID                 string
	Seat               int
	Name               string
	BuyIn              int
	ChipCount          int
	Acted              bool
	Action             ActionType // represents player's current action
	Folded             bool
	AllIn              bool
	ChipsInPot         int
	ChipsInRound       int
	SittingOut         bool
	SittingOutNextHand bool

	// Game play stuff
	Cards  []cards.Card `json:"-"`
	Reveal []cards.Card
}

func NewPlayer(id string, seat int) *Player {
	return &Player{
		ID:   id,
		Seat: seat,
	}
}

func (p *Player) contribute(chips int) {
	amount := chips
	if p.ChipCount <= amount {
		amount = p.ChipCount
		p.AllIn = true
	}
	p.ChipsInPot += amount
	p.ChipsInRound += amount
	p.ChipCount -= amount
}
