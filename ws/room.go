package ws

import (
	"fmt"
	"net"
	"sort"

	"github.com/rs/xid"
	"github.com/screego/server/config"
	"github.com/screego/server/turn"
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
	sessionCreatedTotal.Inc()

	iceHost := []outgoing.ICEServer{}
	iceClient := []outgoing.ICEServer{}
	switch r.Mode {
	case ConnectionLocal:
	case ConnectionSTUN:
		iceHost = []outgoing.ICEServer{{URLs: rooms.addresses("stun", false)}}
		iceClient = []outgoing.ICEServer{{URLs: rooms.addresses("stun", false)}}
	case ConnectionTURN:
		clientAccount := &turn.Account{
			Id: client,
			IP: r.Users[host].Addr,
		}
		hostAccount := &turn.Account{
			Id: host,
			IP: r.Users[client].Addr,
		}
		// FIXME handles error
		err := rooms.turnServer.AcceptAccounts(clientAccount, hostAccount)
		if err != nil {
			panic(err)
		}
		iceHost = []outgoing.ICEServer{{
			URLs:       rooms.addresses("turn", true),
			Credential: hostAccount.Credential,
			Username:   hostAccount.Username,
		}}
		iceClient = []outgoing.ICEServer{{
			URLs:       rooms.addresses("turn", true),
			Credential: clientAccount.Credential,
			Username:   clientAccount.Username,
		}}

	}
	r.Users[host].Write <- outgoing.HostSession{Peer: client, ID: id, ICEServers: iceHost}
	r.Users[client].Write <- outgoing.ClientSession{Peer: host, ID: id, ICEServers: iceClient}
}

// Rooms STUN/TURN addresses
// prefix is `stun` or `turn`
func (r *Rooms) addresses(prefix string, tcp bool) (result []string) {
	if r.config.ExternalIPV4 != nil {
		result = append(result, fmt.Sprintf("%s:%s:%d", prefix, r.config.ExternalIPV4.String(), r.turnServer.Port()))
		if tcp {
			result = append(result, fmt.Sprintf("%s:%s:%d?transport=tcp", prefix, r.config.ExternalIPV4.String(), r.turnServer.Port()))
		}
	}
	if r.config.ExternalIPV6 != nil {
		result = append(result, fmt.Sprintf("%s:[%s]:%d", prefix, r.config.ExternalIPV6.String(), r.turnServer.Port()))
		if tcp {
			result = append(result, fmt.Sprintf("%s:[%s]:%d?transport=tcp", prefix, r.config.ExternalIPV6.String(), r.turnServer.Port()))
		}
	}
	return
}

func (r *Room) closeSession(rooms *Rooms, id xid.ID) {
	if r.Mode == ConnectionTURN {
		session := r.Sessions[id]
		rooms.turnServer.RevokeAccounts(session.Host, session.Client)
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
	Close     chan<- string
}
