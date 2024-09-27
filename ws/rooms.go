package ws

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/xid"
	"github.com/rs/zerolog/log"
	"github.com/screego/server/auth"
	"github.com/screego/server/config"
	"github.com/screego/server/turn"
	"github.com/screego/server/util"
)

func NewRooms(tServer turn.Server, users *auth.Users, conf config.Config) *Rooms {
	return &Rooms{
		Rooms:      map[string]*Room{},
		Incoming:   make(chan ClientMessage),
		connected:  map[xid.ID]bool{},
		turnServer: tServer,
		users:      users,
		config:     conf,
		r:          rand.New(rand.NewSource(time.Now().Unix())),
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
	turnServer turn.Server
	Rooms      map[string]*Room
	Incoming   chan ClientMessage
	upgrader   websocket.Upgrader
	users      *auth.Users
	config     config.Config
	r          *rand.Rand
	connected  map[xid.ID]bool
}

func (r *Rooms) RandUserName() string {
	return util.NewUserName(r.r)
}

func (r *Rooms) RandRoomName() string {
	return util.NewRoomName(r.r)
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
	r.Incoming <- ClientMessage{Info: c.info, Incoming: Connected{}, SkipConnectedCheck: true}

	go c.startReading(time.Second * 20)
	go c.startWriteHandler(time.Second * 5)
}

func (r *Rooms) Start() {
	for msg := range r.Incoming {
		if !msg.SkipConnectedCheck && !r.connected[msg.Info.ID] {
			log.Debug().Interface("event", fmt.Sprintf("%T", msg.Incoming)).Interface("payload", msg.Incoming).Msg("WebSocket Ignore")
			continue
		}

		if err := msg.Incoming.Execute(r, msg.Info); err != nil {
			dis := Disconnected{Code: websocket.CloseNormalClosure, Reason: err.Error()}
			dis.executeNoError(r, msg.Info)
		}
	}
}

func (r *Rooms) closeRoom(roomID string) {
	room, ok := r.Rooms[roomID]
	if !ok {
		return
	}
	usersLeftTotal.Add(float64(len(room.Users)))
	for id := range room.Sessions {
		room.closeSession(r, id)
	}

	delete(r.Rooms, roomID)
	roomsClosedTotal.Inc()
}
