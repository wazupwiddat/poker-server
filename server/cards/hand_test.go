package cards_test

import (
	"encoding/json"
	"math/rand"
	"strings"
	"testing"

	"github.com/wazupwiddat/poker-server/server/cards"
)

type testPair struct {
	cards       []cards.Card
	arrangement []cards.Card
	ranking     cards.Ranking
	description string
}

var tests = []testPair{
	{
		Cards("As", "Ks", "Qs", "2s", "2c", "2h", "2d"),
		Cards("2s", "2c"),
		cards.FourOfAKind,
		"four of a kind twos",
	},
	{
		Cards("Ks", "Qs", "Js", "As", "9d"),
		Cards("As", "Ks", "Qs", "Js", "9d"),
		cards.HighCard,
		"high card ace high",
	},
	{
		Cards("Ks", "Qh", "Qs", "Js", "9d"),
		Cards("Qh", "Qs", "Ks", "Js", "9d"),
		cards.Pair,
		"pair of queens",
	},
	{
		Cards("2s", "Qh", "Qs", "Js", "2d"),
		Cards("Qh", "Qs", "2s", "2d", "Js"),
		cards.TwoPair,
		"two pair queens and twos",
	},
	{
		Cards("6s", "Qh", "Ks", "6h", "6d"),
		Cards("6s", "6h", "6d", "Ks", "Qh"),
		cards.ThreeOfAKind,
		"three of a kind sixes",
	},
	{
		Cards("Ks", "Qs", "Js", "As", "Td"),
		Cards("As", "Ks", "Qs", "Js", "Td"),
		cards.Straight,
		"straight ace high",
	},
	{
		Cards("2s", "3s", "4s", "As", "5d"),
		Cards("5d", "4s", "3s", "2s", "As"),
		cards.Straight,
		"straight five high",
	},
	{
		Cards("7s", "4s", "5s", "3s", "2s"),
		Cards("7s", "5s", "4s", "3s", "2s"),
		cards.Flush,
		"flush seven high",
	},
	{
		Cards("7s", "7d", "3s", "3d", "7h"),
		Cards("7s", "7d", "7h", "3s", "3d"),
		cards.FullHouse,
		"full house sevens full of threes",
	},
	{
		Cards("7s", "7d", "3s", "7c", "7h"),
		Cards("7s", "7d", "7c", "7h", "3s"),
		cards.FourOfAKind,
		"four of a kind sevens",
	},
	{
		Cards("Ks", "Qs", "Js", "Ts", "9s"),
		Cards("Ks", "Qs", "Js", "Ts", "9s"),
		cards.StraightFlush,
		"straight flush king high",
	},
	{
		Cards("As", "5s", "4s", "3s", "2s"),
		Cards("5s", "4s", "3s", "2s", "As"),
		cards.StraightFlush,
		"straight flush five high",
	},
	{
		Cards("As", "Ks", "Qs", "Js", "Ts"),
		Cards("As", "Ks", "Qs", "Js", "Ts"),
		cards.RoyalFlush,
		"royal flush",
	},
	{
		Cards("As", "Ks", "Qs", "2s", "2c", "2h", "2d"),
		Cards("2s", "2c", "2h", "2d", "As"),
		cards.FourOfAKind,
		"four of a kind twos",
	},
}

func TestHands(t *testing.T) {
	for _, test := range tests {
		h := cards.NewHand(test.cards)
		if h.Ranking() != test.ranking {
			t.Fatalf("expected %v got %v", test.ranking, h.Ranking())
		}
		for i := 0; i < 5; i++ {
			actual, expected := h.Cards()[i], test.arrangement[i]
			if actual.Rank() != expected.Rank() || actual.Suit() != expected.Suit() {
				t.Fatalf("expected %v got %v", expected, actual)
			}
		}
		if test.description != h.Description() {
			t.Fatalf("expected \"%v\" got \"%v\"", test.description, h.Description())
		}
	}
}

type equality int

const (
	greaterThan equality = iota
	lessThan
	equalTo
)

type testEquality struct {
	cards1 []cards.Card
	cards2 []cards.Card
	e      equality
}

var equalityTests = []testEquality{
	{
		Cards("As", "5s", "4s", "3s", "2s"),
		Cards("Ks", "Kc", "Kh", "Jd", "Js"),
		greaterThan,
	},
	{
		Cards("Ts", "9h", "8d", "7c", "6s", "2h", "3s"),
		Cards("Ts", "9h", "8d", "7c", "6s", "Ah", "Ks"),
		equalTo,
	},
	{
		Cards("Js", "Jh", "8d", "7c", "6s", "5h", "8s"),
		Cards("Js", "Jh", "8d", "7c", "6s", "5h", "5s"),
		greaterThan,
	},
}

func TestCompareHands(t *testing.T) {
	for _, test := range equalityTests {
		h1 := cards.NewHand(test.cards1)
		h2 := cards.NewHand(test.cards2)
		compareTo := h1.CompareTo(h2)

		switch test.e {
		case greaterThan:
			if compareTo <= 0 {
				t.Errorf("expected %v to be greater than %v", h1, h2)
			}
		case lessThan:
			if compareTo >= 0 {
				t.Errorf("expected %v to be less than %v", h1, h2)
			}
		case equalTo:
			if compareTo != 0 {
				t.Errorf("expected %v to be equal to %v", h1, h2)
			}
		}
	}
}

type testOptionsPairs struct {
	cards       []cards.Card
	arrangement []cards.Card
	options     []func(*cards.Config)
	ranking     cards.Ranking
	description string
}

var optTests = []testOptionsPairs{
	{
		Cards("Ks", "Qs", "Js", "As", "9s"),
		Cards("As", "Ks", "Qs", "Js", "9s"),
		[]func(*cards.Config){cards.Low},
		cards.Flush,
		"flush ace high",
	},
	{
		Cards("7h", "6h", "5s", "4s", "2s", "3s"),
		Cards("6h", "5s", "4s", "3s", "2s"),
		[]func(*cards.Config){cards.AceToFiveLow},
		cards.HighCard,
		"high card six high",
	},
	{
		Cards("Ah", "6h", "5s", "4s", "2s", "Ks"),
		Cards("6h", "5s", "4s", "2s", "Ah"),
		[]func(*cards.Config){cards.AceToFiveLow},
		cards.HighCard,
		"high card six high",
	},
}

func TestHandsWithOptions(t *testing.T) {
	for _, test := range optTests {
		h := cards.NewHand(test.cards, test.options...)
		if h.Ranking() != test.ranking {
			t.Fatalf("expected %v got %v", test.ranking, h.Ranking())
		}
		for i := 0; i < 5; i++ {
			actual, expected := h.Cards()[i], test.arrangement[i]
			if actual.Rank() != expected.Rank() || actual.Suit() != expected.Suit() {
				t.Fatalf("expected %v got %v", expected, actual)
			}
		}
		if test.description != h.Description() {
			t.Fatalf("expected \"%v\" got \"%v\"", test.description, h.Description())
		}
	}
}

func TestBlanks(t *testing.T) {
	c := []cards.Card{cards.AceSpades}
	h := cards.NewHand(c)
	if h.Ranking() != cards.HighCard {
		t.Fatal("blank card error")
	}

	c = []cards.Card{cards.FiveSpades, cards.FiveClubs}
	h = cards.NewHand(c)
	if h.Ranking() != cards.Pair {
		t.Fatal("blank card error")
	}
}

func TestDeck(t *testing.T) {
	r := rand.New(rand.NewSource(0))
	deck := cards.NewDealer(r).Deck()
	if deck.Pop() == deck.Pop() {
		t.Fatal("Two Pop() calls should never return the same result")
	}
	l := len(deck.Cards)
	if l != 50 {
		t.Fatalf("After Pop() deck len = %d; want %d", l, 50)
	}
}

func TestHandJSON(t *testing.T) {
	jsonStr := `{"ranking":10,"cards":["A♠","K♠","Q♠","J♠","T♠"],"description":"royal flush","config":{"sorting":1,"ignoreStraights":false,"ignoreFlushes":false,"aceIsLow":false}}`
	h := &cards.Hand{}
	if err := json.Unmarshal([]byte(jsonStr), h); err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(h)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != jsonStr {
		t.Fatalf("expected json %s but got %s", jsonStr, string(b))
	}
}

func BenchmarkHandCreation(b *testing.B) {
	r := rand.New(rand.NewSource(0))
	c := cards.NewDealer(r).Deck().PopMulti(7)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cards.NewHand(c)
	}
}

const (
	deck1Str = "J♦ 7♥ 2♥ 3♣ 5♥ 5♦ 2♦ 3♦ Q♦ 9♠ A♣ 9♣ T♠ 7♦ J♥ 4♦ A♦ J♣ K♣ 9♥ T♦ 2♠ 6♣ 2♣ 6♦ 7♣ 8♣ K♠ 8♠ 6♠ 5♠ 6♥ Q♠ 5♣ Q♣ Q♥ 4♣ 3♠ A♠ 8♦ K♦ 9♦ 4♥ K♥ 8♥ T♥ 3♥ A♥ T♣ J♠ 7♠ 4♠"
	deck2Str = "A♥ 4♥ 3♣ 2♥ 9♥ 9♠ 3♦ 8♥ 2♦ 6♣ 2♣ T♥ K♦ 9♣ 7♣ 7♠ 6♦ J♣ 8♠ J♥ Q♣ 6♥ T♠ A♦ 8♦ 8♣ J♦ 5♦ Q♦ A♣ 2♠ T♦ K♣ A♠ Q♠ 6♠ 4♦ 5♠ Q♥ 3♠ K♥ 7♦ 5♥ 4♣ 3♥ 9♦ 7♥ J♠ K♠ 4♠ 5♣ T♣"
	deck3Str = "Q♠ 5♦ 5♣ 4♦ Q♣ 4♥ 7♥ Q♥ T♦ 2♦ 4♠ T♣ J♣ A♣ 8♥ 3♠ 7♣ 9♥ 8♣ 9♣ 6♦ 6♣ 8♦ A♦ K♥ J♦ 7♠ 2♥ 7♦ 3♥ A♥ 9♠ K♣ 2♣ 8♠ 5♠ 6♥ T♠ T♥ A♠ 3♦ 9♦ 6♠ K♦ J♥ K♠ 4♣ 3♣ J♠ Q♦ 5♥ 2♠"
	deck4Str = "6♠ J♠ J♣ 8♣ Q♥ A♣ T♦ T♠ Q♦ 5♠ Q♠ 9♠ 4♦ 7♦ 3♥ 4♣ 8♥ A♥ 6♣ 7♣ T♣ 7♠ K♣ 3♠ 4♥ K♥ 9♥ 5♥ 6♦ 3♣ 3♦ 7♥ 2♣ 6♥ T♥ 9♣ J♦ 9♦ 2♦ K♦ 8♠ K♠ 4♠ J♥ Q♣ 2♠ 2♥ 5♦ 8♦ A♦ 5♣ A♠"
	deck5Str = "5♥ 6♥ 6♣ 3♠ T♣ Q♣ 5♦ A♦ 5♣ J♦ 9♦ 9♣ A♣ 8♠ K♥ 8♦ 7♣ K♣ T♦ 2♥ Q♦ 5♠ Q♥ K♠ 8♣ 4♥ 3♣ K♦ 2♣ T♠ T♥ 8♥ 4♣ Q♠ 4♦ A♥ 3♦ 6♠ 9♠ A♠ 2♠ 7♠ 2♦ 9♥ 4♠ 6♦ 3♥ J♣ 7♦ J♥ 7♥ J♠"
)

func Deck1() *cards.Deck {
	return parseDeck(deck1Str)
}

func Deck2() *cards.Deck {
	return parseDeck(deck2Str)
}

func Deck3() *cards.Deck {
	return parseDeck(deck3Str)
}

func Deck4() *cards.Deck {
	return parseDeck(deck4Str)
}

func Deck5() *cards.Deck {
	return parseDeck(deck5Str)
}

func parseDeck(s string) *cards.Deck {
	cds := []cards.Card{}
	for _, cardStr := range strings.Split(s, " ") {
		temp := cards.AceSpades
		c := &temp
		if err := c.UnmarshalText([]byte(cardStr)); err != nil {
			panic(err)
		}
		cds = append(cds, *c)
	}
	return &cards.Deck{Cards: cds}
}

// Cards takes a list of strings that have the format "4s", "Tc",
// "Ah" instead of the hand.Card String() format "4♠", "T♣", "A♥"
// for ease of testing.  If a string is invalid Cards panics,
// otherwise it returns a list of the corresponding cards.
func Cards(list ...string) []cards.Card {
	cards := []cards.Card{}
	for _, s := range list {
		cards = append(cards, card(s))
	}
	return cards
}

// Dealer returns a hand.Dealer that generates decks that will pop
// cards in the order of the cards given.
func Dealer(cds []cards.Card) cards.Dealer {
	return &deck{cards: cds}
}

type deck struct {
	cards []cards.Card
}

func (d deck) Deck() *cards.Deck {
	// copy cards
	c := make([]cards.Card, len(d.cards))
	copy(c, d.cards)

	// reverse cards
	for i, j := 0, len(c)-1; i < j; i, j = i+1, j-1 {
		c[i], c[j] = c[j], c[i]
	}
	return &cards.Deck{Cards: c}
}

func card(s string) cards.Card {
	if len(s) != 2 {
		panic("jokertest: card string must be two characters")
	}

	rank, ok := rankMap[s[:1]]
	if !ok {
		panic("jokertest: rank not found")
	}

	suit, ok := suitMap[s[1:]]
	if !ok {
		panic("jokertest: suit not found")
	}

	for _, c := range cards.Cards() {
		if rank == c.Rank() && suit == c.Suit() {
			return c
		}
	}
	panic("jokertest: card not found")
}

var (
	rankMap = map[string]cards.Rank{
		"A": cards.Ace,
		"K": cards.King,
		"Q": cards.Queen,
		"J": cards.Jack,
		"T": cards.Ten,
		"9": cards.Nine,
		"8": cards.Eight,
		"7": cards.Seven,
		"6": cards.Six,
		"5": cards.Five,
		"4": cards.Four,
		"3": cards.Three,
		"2": cards.Two,
	}

	suitMap = map[string]cards.Suit{
		"s": cards.Spades,
		"h": cards.Hearts,
		"d": cards.Diamonds,
		"c": cards.Clubs,
	}
)
