package tables

// import (
// 	"testing"

// 	. "github.com/smartystreets/goconvey/convey"
// )

// func createTestTable() *Table {
// 	testTable := NewTable("1")
// 	go testTable.HandleTableUpdates()

// 	testTable.AddPlayerToSeat("sid1", 0)
// 	testTable.AddPlayerToSeat("sid2", 1)
// 	return testTable
// }

// func TestAction(t *testing.T) {
// 	Convey("Given a new table", t, func() {
// 		testTable := createTestTable()
// 		game := NewCardGame(testTable)
// 		testTable.game = game
// 		Convey("Game can't start until players have bought in", func() {
// 			game.Updates(func(gameUpdate GameUpdate) {
// 				So(gameUpdate.Button, ShouldEqual, -1)
// 				So(testTable.Seats[0].SittingOut, ShouldBeTrue)
// 				So(testTable.Seats[1].SittingOut, ShouldBeTrue)
// 			})
// 			game.StartRound()
// 		})
// 		Convey("Two player with enough to play", func() {
// 			testTable.AddChipsToPlayer("sid1", 200)
// 			testTable.AddChipsToPlayer("sid2", 200)
// 			game.Updates(func(gameUpdate GameUpdate) {
// 				So(gameUpdate.Button, ShouldEqual, 0)
// 				So(gameUpdate.Round, ShouldEqual, PreFlop)
// 				So(testTable.Seats[0].ChipCount, ShouldEqual, 198)
// 				So(testTable.Seats[0].ChipsInPot, ShouldEqual, 2)
// 				So(testTable.Seats[0].ChipsInRound, ShouldEqual, 2)
// 				So(testTable.Seats[1].ChipCount, ShouldEqual, 199)
// 				So(testTable.Seats[1].ChipsInPot, ShouldEqual, 1)
// 				So(testTable.Seats[1].ChipsInRound, ShouldEqual, 1)
// 				So(gameUpdate.Active, ShouldResemble, testTable.Seats[1])
// 				So(gameUpdate.Cost, ShouldEqual, 2)
// 				So(gameUpdate.Cards, ShouldBeNil)
// 			})
// 			game.StartRound()
// 			Convey("Action Call", func() {
// 				game.Updates(func(gameUpdate GameUpdate) {
// 					So(gameUpdate.Button, ShouldEqual, 0)
// 					So(gameUpdate.Round, ShouldEqual, PreFlop)
// 					So(testTable.Seats[0].ChipCount, ShouldEqual, 198)
// 					So(testTable.Seats[0].ChipsInPot, ShouldEqual, 2)
// 					So(testTable.Seats[0].ChipsInRound, ShouldEqual, 2)
// 					So(testTable.Seats[1].ChipCount, ShouldEqual, 198)
// 					So(testTable.Seats[1].ChipsInPot, ShouldEqual, 2)
// 					So(testTable.Seats[1].ChipsInRound, ShouldEqual, 2)
// 					So(gameUpdate.Active, ShouldResemble, testTable.Seats[0])
// 					So(gameUpdate.Cost, ShouldEqual, 2)
// 					So(gameUpdate.Cards, ShouldBeNil)
// 				})
// 				testTable.Act(Action{
// 					Type:  Call,
// 					Chips: 0,
// 				})
// 			})
// 			Convey("Action Raise", func() {
// 				game.Updates(func(gameUpdate GameUpdate) {
// 					So(gameUpdate.Button, ShouldEqual, 0)
// 					So(gameUpdate.Round, ShouldEqual, PreFlop)
// 					So(testTable.Seats[0].ChipCount, ShouldEqual, 198)
// 					So(testTable.Seats[0].ChipsInPot, ShouldEqual, 2)
// 					So(testTable.Seats[0].ChipsInRound, ShouldEqual, 2)
// 					So(testTable.Seats[1].ChipCount, ShouldEqual, 188)
// 					So(testTable.Seats[1].ChipsInPot, ShouldEqual, 12)
// 					So(testTable.Seats[1].ChipsInRound, ShouldEqual, 12)
// 					So(gameUpdate.Active, ShouldResemble, testTable.Seats[0])
// 					So(gameUpdate.Cost, ShouldEqual, 12)
// 					So(gameUpdate.Cards, ShouldBeNil)
// 				})
// 				testTable.Act(Action{
// 					Type:  Raise,
// 					Chips: 10,
// 				})
// 			})
// 			Convey("Action Folds", func() {
// 				game.Updates(func(gameUpdate GameUpdate) {
// 					So(gameUpdate.Button, ShouldEqual, 1)
// 					So(gameUpdate.Round, ShouldEqual, PreFlop)
// 					So(testTable.Seats[0].ChipCount, ShouldEqual, 200)
// 					So(testTable.Seats[0].ChipsInPot, ShouldEqual, 1)
// 					So(testTable.Seats[0].ChipsInRound, ShouldEqual, 1)
// 					So(testTable.Seats[1].ChipCount, ShouldEqual, 197)
// 					So(testTable.Seats[1].ChipsInPot, ShouldEqual, 2)
// 					So(testTable.Seats[1].ChipsInRound, ShouldEqual, 2)
// 					So(gameUpdate.Active, ShouldResemble, testTable.Seats[0])
// 					So(gameUpdate.Cost, ShouldEqual, 2)
// 					So(gameUpdate.Cards, ShouldBeNil)
// 				})
// 				testTable.Act(Action{
// 					Type:  Fold,
// 					Chips: 0,
// 				})
// 			})
// 		})
// 		Convey("Player joining mid round has to wait", func() {
// 			testTable.AddChipsToPlayer("sid1", 200)
// 			testTable.AddChipsToPlayer("sid2", 200)
// 			game.Updates(func(gameUpdate GameUpdate) {
// 				So(gameUpdate.Button, ShouldEqual, 0)
// 				So(gameUpdate.Round, ShouldEqual, PreFlop)
// 				So(testTable.Seats[0].ChipCount, ShouldEqual, 198)
// 				So(testTable.Seats[0].ChipsInPot, ShouldEqual, 2)
// 				So(testTable.Seats[0].ChipsInRound, ShouldEqual, 2)
// 				So(testTable.Seats[1].ChipCount, ShouldEqual, 199)
// 				So(testTable.Seats[1].ChipsInPot, ShouldEqual, 1)
// 				So(testTable.Seats[1].ChipsInRound, ShouldEqual, 1)
// 				So(gameUpdate.Active, ShouldResemble, testTable.Seats[1])
// 				So(gameUpdate.Cost, ShouldEqual, 2)
// 				So(gameUpdate.Cards, ShouldBeNil)
// 			})
// 			game.StartRound()
// 			game.Updates(func(gameUpdate GameUpdate) {
// 				So(gameUpdate.Button, ShouldEqual, 0)
// 				So(gameUpdate.Round, ShouldEqual, PreFlop)
// 				So(testTable.Seats[0].ChipCount, ShouldEqual, 198)
// 				So(testTable.Seats[0].ChipsInPot, ShouldEqual, 2)
// 				So(testTable.Seats[0].ChipsInRound, ShouldEqual, 2)
// 				So(testTable.Seats[1].ChipCount, ShouldEqual, 198)
// 				So(testTable.Seats[1].ChipsInPot, ShouldEqual, 2)
// 				So(testTable.Seats[1].ChipsInRound, ShouldEqual, 2)
// 				So(gameUpdate.Active, ShouldResemble, testTable.Seats[0])
// 				So(gameUpdate.Cost, ShouldEqual, 2)
// 				So(gameUpdate.Cards, ShouldBeNil)
// 			})
// 			testTable.Act(Action{
// 				Type:  Call,
// 				Chips: 0,
// 			})
// 			game.Updates(func(gameUpdate GameUpdate) {
// 				So(gameUpdate.Button, ShouldEqual, 0)
// 				So(gameUpdate.Round, ShouldEqual, Flop)
// 				So(testTable.Seats[0].ChipCount, ShouldEqual, 198)
// 				So(testTable.Seats[0].ChipsInPot, ShouldEqual, 2)
// 				So(testTable.Seats[0].ChipsInRound, ShouldEqual, 0)
// 				So(testTable.Seats[1].ChipCount, ShouldEqual, 198)
// 				So(testTable.Seats[1].ChipsInPot, ShouldEqual, 2)
// 				So(testTable.Seats[1].ChipsInRound, ShouldEqual, 0)
// 				So(gameUpdate.Active, ShouldResemble, testTable.Seats[1])
// 				So(gameUpdate.Cost, ShouldEqual, 0)
// 				So(gameUpdate.Cards, ShouldNotBeNil)
// 			})
// 			testTable.Act(Action{
// 				Type:  Check,
// 				Chips: 0,
// 			})
// 			testTable.AddPlayerToSeat("sid3", 2)
// 			testTable.AddChipsToPlayer("sid3", 200)
// 			game.Updates(func(gameUpdate GameUpdate) {
// 				So(gameUpdate.Button, ShouldEqual, 0)
// 				So(gameUpdate.Round, ShouldEqual, Flop)
// 				So(testTable.Seats[0].ChipCount, ShouldEqual, 198)
// 				So(testTable.Seats[0].ChipsInPot, ShouldEqual, 2)
// 				So(testTable.Seats[0].ChipsInRound, ShouldEqual, 0)
// 				So(testTable.Seats[1].ChipCount, ShouldEqual, 198)
// 				So(testTable.Seats[1].ChipsInPot, ShouldEqual, 2)
// 				So(testTable.Seats[1].ChipsInRound, ShouldEqual, 0)
// 				So(gameUpdate.Active, ShouldResemble, testTable.Seats[0])
// 				So(gameUpdate.Cost, ShouldEqual, 0)
// 				So(gameUpdate.Cards, ShouldNotBeNil)
// 			})
// 			testTable.Act(Action{
// 				Type:  Check,
// 				Chips: 0,
// 			})
// 		})
// 	})
// }
