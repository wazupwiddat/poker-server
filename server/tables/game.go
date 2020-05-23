package tables

import (
	"fmt"

	"github.com/wazupwiddat/poker-server/server/cards"
)

type Action struct {
	Type  ActionType
	Chips int
}

func (a Action) String() string {
	return fmt.Sprintf("Action: %s, Chips: %d", a.Type.String(), a.Chips)
}

type ActionType int

const (
	Fold ActionType = iota
	Check
	Call
	Bet
	Raise
	AllIn
)

func (at ActionType) String() string {
	return []string{"Fold", "Check", "Call", "Bet", "Raise", "AllIn"}[at]
}

type SidePot struct {
	Contesting []*Player
	Chips      int
}

type GameUpdate struct {
	Active     *Player
	Round      Round
	Cards      []cards.Card
	Button     int
	Cost       int
	Pots       []*SidePot
	LastWinner *cards.Hand
	Winners    []*Player
}

type CardGame interface {
	StartRound()
	Act(*Player, Action) error
	ForceUpdate() // this should execute the update function below
	Updates(func(GameUpdate))
	Stop()
}

func NewCardGame(table *Table) CardGame {
	switch table.Options.Variant {
	case TexasHoldem:
		return &TexasHoldemCardGame{
			table:  table,
			button: -1,
		}
	case OmahaHi:
		return &OmahaHiCardGame{
			table: table,
		}
	default:
		return &TexasHoldemCardGame{
			table:  table,
			button: -1,
		}
	}
}
