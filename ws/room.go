package ws

import (
	"fmt"
	"net"
	"sort"

	"github.com/rs/xid"
	"github.com/screego/server/config"
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

func (r *Room) newSession(host, client xid.ID, rooms *Rooms, v4, v6 net.IP) {
	id := xid.New()
	r.Sessions[id] = &RoomSession{
		Host:   host,
		Client: client,
	}
	sessionCreatedTotal.Inc()

	iceHost := []outgoing.ICEServer{}
	iceClient := []outgoing.ICEServer{}
	switch r.Mode {
	case ConnectionLocal:
	case ConnectionSTUN:
		iceHost = []outgoing.ICEServer{{URLs: rooms.addresses("stun", v4, v6, false)}}
		iceClient = []outgoing.ICEServer{{URLs: rooms.addresses("stun", v4, v6, false)}}
	case ConnectionTURN:
		hostName, hostPW := rooms.turnServer.Credentials(id.String()+"host", r.Users[host].Addr)
		clientName, clientPW := rooms.turnServer.Credentials(id.String()+"client", r.Users[client].Addr)
		iceHost = []outgoing.ICEServer{{
			URLs:       rooms.addresses("turn", v4, v6, true),
			Credential: hostPW,
			Username:   hostName,
		}}
		iceClient = []outgoing.ICEServer{{
			URLs:       rooms.addresses("turn", v4, v6, true),
			Credential: clientPW,
			Username:   clientName,
		}}
	}
	r.Users[host].Write <- outgoing.HostSession{Peer: client, ID: id, ICEServers: iceHost}
	r.Users[client].Write <- outgoing.ClientSession{Peer: host, ID: id, ICEServers: iceClient}
}

func (r *Rooms) addresses(prefix string, v4, v6 net.IP, tcp bool) (result []string) {
	if v4 != nil {
		result = append(result, fmt.Sprintf("%s:%s:%s", prefix, v4.String(), r.config.TurnPort))
		if tcp {
			result = append(result, fmt.Sprintf("%s:%s:%s?transport=tcp", prefix, v4.String(), r.config.TurnPort))
		}
	}
	if v6 != nil {
		result = append(result, fmt.Sprintf("%s:[%s]:%s", prefix, v6.String(), r.config.TurnPort))
		if tcp {
			result = append(result, fmt.Sprintf("%s:[%s]:%s?transport=tcp", prefix, v6.String(), r.config.TurnPort))
		}
	}
	return
}

func (r *Room) closeSession(rooms *Rooms, id xid.ID) {
	if r.Mode == ConnectionTURN {
		rooms.turnServer.Disallow(id.String() + "host")
		rooms.turnServer.Disallow(id.String() + "client")
	}
	delete(r.Sessions, id)
	sessionClosedTotal.Inc()
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
}
