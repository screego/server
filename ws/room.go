package ws

import (
	"net"
	"sort"

	"github.com/rs/xid"
	"github.com/screego/server/config"
	"github.com/screego/server/util"
	"github.com/screego/server/ws/outgoing"
)

type ConnectionMode string

const (
	ConnectionLocal ConnectionMode = "local"
	ConnectionSTUN  ConnectionMode = "stun"
	ConnectionTURN  ConnectionMode = config.AuthModeTurn
)

type Room struct {
	ID                string
	CloseOnOwnerLeave bool
	Mode              ConnectionMode
	Users             map[xid.ID]*User
	Sessions          map[xid.ID]*RoomSession
}

const (
	CloseOwnerLeft = "Owner Left"
	CloseDone      = "Read End"
)

func (r *Room) newSession(host, client xid.ID, rooms *Rooms) {
	id := xid.New()
	r.Sessions[id] = &RoomSession{
		Host:   host,
		Client: client,
	}

	iceHost := []outgoing.ICEServer{}
	iceClient := []outgoing.ICEServer{}
	switch r.Mode {
	case ConnectionLocal:
	case ConnectionSTUN:
		iceHost = []outgoing.ICEServer{{URLs: []string{rooms.turnServer.StunAddress}}}
		iceClient = []outgoing.ICEServer{{URLs: []string{rooms.turnServer.StunAddress}}}
	case ConnectionTURN:
		hostPW := util.RandString(20)
		clientPW := util.RandString(20)
		hostName := id.String() + "host"
		rooms.turnServer.Allow(hostName, hostPW, r.Users[host].Addr)
		clientName := id.String() + "client"
		rooms.turnServer.Allow(clientName, clientPW, r.Users[client].Addr)
		iceHost = []outgoing.ICEServer{{
			URLs: []string{
				rooms.turnServer.TurnAddress,
				rooms.turnServer.TurnAddress + "?transport=tcp",
			},
			Credential: hostPW,
			Username:   hostName,
		}}
		iceClient = []outgoing.ICEServer{{
			URLs: []string{
				rooms.turnServer.TurnAddress,
				rooms.turnServer.TurnAddress + "?transport=tcp",
			},
			Credential: clientPW,
			Username:   clientName,
		}}

	}
	r.Users[host].Write <- outgoing.HostSession{Peer: client, ID: id, ICEServers: iceHost}
	r.Users[client].Write <- outgoing.ClientSession{Peer: host, ID: id, ICEServers: iceClient}
}

type RoomSession struct {
	Host   xid.ID
	Client xid.ID
}

func (r *Room) notifyInfoChanged() {
	for _, current := range r.Users {
		users := []outgoing.User{}
		for _, user := range r.Users {
			users = append(users, outgoing.User{
				ID:        user.ID,
				Name:      user.Name,
				Streaming: user.Streaming,
				You:       current == user,
				Owner:     user.Owner,
			})
		}

		sort.Slice(users, func(i, j int) bool {
			left := users[i]
			right := users[j]

			if left.Owner != right.Owner {
				return left.Owner
			}

			if left.Streaming != right.Streaming {
				return left.Streaming
			}

			return left.Name < right.Name
		})

		current.Write <- outgoing.Room{
			ID:    r.ID,
			Users: users,
		}
	}
}

type User struct {
	ID        xid.ID
	Addr      net.IP
	Name      string
	Streaming bool
	Owner     bool
	Write     chan<- outgoing.Message
	Close     chan<- string
}
