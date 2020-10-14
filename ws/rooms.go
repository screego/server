package ws

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"github.com/screego/server/auth"
	"github.com/screego/server/config"
	"github.com/screego/server/turn"
)

func NewRooms(tServer *turn.Server, users *auth.Users, conf config.Config) *Rooms {
	return &Rooms{
		Rooms:      map[string]*Room{},
		Incoming:   make(chan ClientMessage),
		turnServer: tServer,
		users:      users,
		config:     conf,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("origin")
				u, err := url.Parse(origin)
				if err != nil {
					return false
				}
				if u.Host == r.Host {
					return true
				}
				return conf.CheckOrigin(origin)
			},
		},
	}
}

type Rooms struct {
	turnServer *turn.Server
	Rooms      map[string]*Room
	Incoming   chan ClientMessage
	upgrader   websocket.Upgrader
	users      *auth.Users
	config     config.Config
}

func (r *Rooms) Upgrade(w http.ResponseWriter, req *http.Request) {
	conn, err := r.upgrader.Upgrade(w, req, nil)

	if err != nil {
		log.Debug().Err(err).Msg("Websocket upgrade")
		w.WriteHeader(400)
		_, _ = w.Write([]byte(fmt.Sprintf("Upgrade failed %s", err)))
		return
	}

	user, loggedIn := r.users.CurrentUser(req)
	c := newClient(conn, req, r.Incoming, user, loggedIn, r.config.TrustProxyHeaders)

	go c.startReading(time.Second * 20)
	go c.startWriteHandler(time.Second * 5)
}

func (r *Rooms) Start() {
	for {
		msg := <-r.Incoming
		if err := msg.Incoming.Execute(r, msg.Info); err != nil {
			msg.Info.Close <- err.Error()
		}
	}
}

func (r *Rooms) closeRoom(roomId string) {
	room, ok := r.Rooms[roomId]
	if !ok {
		return
	}
	usersLeftTotal.Add(float64(len(room.Users)))
	sessionClosedTotal.Add(float64(len(room.Sessions)))
	delete(r.Rooms, roomId)
	roomsClosedTotal.Inc()
}
