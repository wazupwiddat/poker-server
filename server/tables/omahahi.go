package tables

type OmahaHiCardGame struct {
	updates func(GameUpdate)
	table   *Table
}

func (c *OmahaHiCardGame) StartRound() {

}

func (c *OmahaHiCardGame) Act(player *Player, act Action) error {
	return nil
}

func (c *OmahaHiCardGame) Updates(f func(GameUpdate)) {
	if f != nil {
		c.updates = f
	}
}

func (c *OmahaHiCardGame) ForceUpdate() {
	// this should call the Updates function passed in
}

func (c *OmahaHiCardGame) Stop() {

}
