package controllers

import (
	"log"
	"net/http"
	"time"

	engineio "github.com/googollee/go-engine.io"
	socketio "github.com/googollee/go-socket.io"
	"github.com/speps/go-hashids"

	"github.com/wazupwiddat/poker-server/server/tables"
)

const (
	DefaultNameSpace = "/"
)

type PokerServer struct {
	server *socketio.Server

	tables         map[string]*tables.Table
	tablesBySocket map[string]*tables.Table // convenience lookup
}

func NewPokerServer(opt *engineio.Options) (*PokerServer, error) {
	s, err := socketio.NewServer(opt)
	if err != nil {
		return nil, err
	}
	t := make(map[string]*tables.Table)
	tbs := make(map[string]*tables.Table)
	return &PokerServer{
		server:         s,
		tables:         t,
		tablesBySocket: tbs,
	}, nil
}

func (p *PokerServer) Serve() error {
	return p.server.Serve()
}

func (p *PokerServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.server.ServeHTTP(w, r)
}

func (p *PokerServer) Close() error {
	return p.server.Close()
}

func (p *PokerServer) Setup() {
	// Default namespace
	p.server.OnConnect(DefaultNameSpace, p.onConnect)
	p.server.OnDisconnect(DefaultNameSpace, p.onDisconnect)
	p.server.OnError(DefaultNameSpace, p.onError)
	p.server.OnEvent(DefaultNameSpace, "createGame", p.onCreateGame)
	p.server.OnEvent(DefaultNameSpace, "joinGame", p.onJoinGame)
	p.server.OnEvent(DefaultNameSpace, "selectSeat", p.onSelectSeat)
	p.server.OnEvent(DefaultNameSpace, "leaveSeat", p.onLeaveSeat)
	p.server.OnEvent(DefaultNameSpace, "username", p.onUsername)
	p.server.OnEvent(DefaultNameSpace, "buyChips", p.onBuyChips)
	p.server.OnEvent(DefaultNameSpace, "rejoinGame", p.onRejoinGame)

	p.server.OnEvent(DefaultNameSpace, "start", p.onStart)
	p.server.OnEvent(DefaultNameSpace, "action", p.onAction)
	p.server.OnEvent(DefaultNameSpace, "showmycards", p.onShowMyCards)
	p.server.OnEvent(DefaultNameSpace, "peekatcards", p.onPeekAtCards)
	p.server.OnEvent(DefaultNameSpace, "sitout", p.onSitout)
	p.server.OnEvent(DefaultNameSpace, "sitin", p.onSitin)

	p.server.OnEvent(DefaultNameSpace, "chat", p.onChat)
}

func (p *PokerServer) onConnect(s socketio.Conn) error {
	log.Printf("client %s connected to %s", s.ID(), s.Namespace())
	return nil
}

func (p *PokerServer) onDisconnect(s socketio.Conn, reason string) {
	log.Printf("client %s disconnected from %s for %s", s.ID(), s.Namespace(), reason)
	// find socket and remove from table
	t, ok := p.tablesBySocket[s.ID()]
	if ok {
		log.Println("removing player from:", t.ID)
		t.RemovePlayerSocket(s.ID())
		delete(p.tablesBySocket, s.ID())
	}

}

func (p *PokerServer) onError(s socketio.Conn, e error) {
	log.Println("Error: ", e)
}

func (p *PokerServer) onCreateGame(s socketio.Conn, gameOpts string) {
	log.Println("onCreateGame:", s.ID())
	log.Println(gameOpts)
	//  generate key, setup table
	gameKey := generateTableJoinKey()

	log.Println("Game key:", gameKey)

	// TODO gameOpts needs to be passed in to NewTable
	table := tables.NewTable(gameKey)
	table.AddPlayerSocket(s)
	p.tables[gameKey] = table
	p.tablesBySocket[s.ID()] = table
	table.OnClose = func(t *tables.Table) {
		delete(p.tables, t.ID)
		delete(p.tablesBySocket, s.ID())
		log.Println("Tabled closed:", t.ID)
	}

	// Handlers will run until the channels close
	go table.HandleTableUpdates()
	go table.HandleChatMessages()
	go table.HandleSavingTableData()
	// go table.Ping()

	log.Println("Table setup and running:", gameKey)
	s.Emit("createGameResponse", gameKey)
}

func (p *PokerServer) onJoinGame(s socketio.Conn, gameKey string) {
	log.Println("onJoinGame", s.ID(), gameKey)

	table, ok := p.tables[gameKey]
	if !ok {
		s.Emit("newJoinError", "no game found for that key")
		return
	}
	table.AddPlayerSocket(s)
	p.tablesBySocket[s.ID()] = table
	s.Emit("newJoinResponse", gameKey)

	table.UpdateAll()
}

func (p *PokerServer) onRejoinGame(s socketio.Conn, gameKey string, seat int, username string) {
	log.Println("rejoinGame", s.ID(), gameKey)

	table, ok := p.tables[gameKey]
	if !ok {
		s.Emit("rejoinGameResponseError", "no game found for that key")
		return
	}

	for i, st := range table.Seats {
		if st == nil {
			continue
		}
		if st.Name == username && i == seat {
			table.AddPlayerSocket(s) // old socket was already removed
			p.tablesBySocket[s.ID()] = table

			st.ID = s.ID()
			table.UpdateAll()

			s.Emit("rejoinGameResponse")
			return
		}
	}
	s.Emit("rejoinGameResponseError", "no user or seat for that table")
}

func (p *PokerServer) onSelectSeat(s socketio.Conn, seat int) {
	log.Println("selectSeat", s.ID(), seat)
	table, ok := p.tablesBySocket[s.ID()]
	if !ok {
		s.Emit("selectSeatErr", "no game found for this connection")
		return
	}
	if added := table.AddPlayerToSeat(s.ID(), seat); !added {
		s.Emit("selectSeatErr", "seat alread in use")
		return
	} // players are 0 based
	s.Emit("selectSeatResponse", seat)
}

func (p *PokerServer) onLeaveSeat(s socketio.Conn) {
	log.Println("leaveSeat", s.ID())
	table, ok := p.tablesBySocket[s.ID()]
	if !ok {
		s.Emit("selectSeatErr", "no game found for this connection")
		return
	}
	table.RemovePlayerFromSeat(s.ID())
	table.RemovePlayerSocket(s.ID())
	s.Emit("leaveSeatResponse")
}

func (p *PokerServer) onUsername(s socketio.Conn, uname string) {
	log.Println("username", s.ID())
	table, ok := p.tablesBySocket[s.ID()]
	if !ok {
		s.Emit("buyChipsErr", "no game found for this connection")
		return
	}
	table.UpdateUsernameToPlayer(s.ID(), uname)
}

func (p *PokerServer) onBuyChips(s socketio.Conn, amt int) {
	log.Println("buyChips", s.ID())
	table, ok := p.tablesBySocket[s.ID()]
	if !ok {
		s.Emit("buyChipsErr", "no game found for this connection")
		return
	}
	table.AddChipsToPlayer(s.ID(), amt)
}

func (p *PokerServer) onStart(s socketio.Conn) {
	log.Println("start", s.ID())
	table, ok := p.tablesBySocket[s.ID()]
	if !ok {
		s.Emit("start", "no game found for this connection")
		return
	}
	table.StartDealing()
}

func (p *PokerServer) onAction(s socketio.Conn, act tables.Action) {
	log.Println("action", s.ID())
	table, ok := p.tablesBySocket[s.ID()]
	if !ok {
		s.Emit("action", "no game found for this connection")
		return
	}
	if err := table.Act(s.ID(), act); err != nil {
		s.Emit("actionError", err.Error())
	}
}

func (p *PokerServer) onPeekAtCards(s socketio.Conn) {
	log.Println("peekAtCards", s.ID())
	table, ok := p.tablesBySocket[s.ID()]
	if !ok {
		s.Emit("showMyCards", "no game found for this connection")
		return
	}
	cards := table.PeekAtCards(s.ID())
	if cards != nil {
		s.Emit("playerCards", cards)
	}
}

func (p *PokerServer) onShowMyCards(s socketio.Conn) {
	log.Println("showMyCards", s.ID())
	table, ok := p.tablesBySocket[s.ID()]
	if !ok {
		s.Emit("showMyCards", "no game found for this connection")
		return
	}
	table.RevealMyCards(s.ID())
}

func (p *PokerServer) onSitout(s socketio.Conn) {
	t, ok := p.tablesBySocket[s.ID()]
	if !ok {
		log.Println("Chat message sent but does not belong to table")
		return
	}

	t.Sitout(s.ID())
}

func (p *PokerServer) onSitin(s socketio.Conn) {
	t, ok := p.tablesBySocket[s.ID()]
	if !ok {
		log.Println("Chat message sent but does not belong to table")
		return
	}

	t.Sitin(s.ID())
}

func (p *PokerServer) onChat(s socketio.Conn, msg string) {
	t, ok := p.tablesBySocket[s.ID()]
	if !ok {
		log.Println("Chat message sent but does not belong to table")
		return
	}
	t.SendChat(s.ID(), msg)
}

func generateTableJoinKey() string {
	hd := hashids.NewData()
	hd.Salt = "somesaltsaltsome"
	hd.MinLength = 6
	h, _ := hashids.NewWithData(hd)
	n := time.Now()
	e, _ := h.Encode([]int{n.Day(), n.Hour(), n.Minute(), n.Second()})
	return e
}
