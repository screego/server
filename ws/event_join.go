package ws

import (
	"errors"
	"fmt"

	"github.com/screego/server/config"
)

func init() {
	register("join", func() Event {
		return &Join{}
	})
}

type Join struct {
	ID       string `json:"id"`
	UserName string `json:"username,omitempty"`
}

func (e *Join) Execute(rooms *Rooms, current ClientInfo) error {
	if rooms.connected[current.ID] != "" {
		return fmt.Errorf("cannot join room, you are already in one")
	}

	room, ok := rooms.Rooms[e.ID]
	if !ok {
		return fmt.Errorf("room with id %s does not exist", e.ID)
	}

	if rooms.config.AuthMode == config.AuthModeAll && !current.Authenticated {
		return errors.New("you need to login")
	}
	name := e.UserName
	if current.Authenticated {
		name = current.AuthenticatedUser
	}
	if name == "" {
		name = rooms.RandUserName()
	}

	room.Users[current.ID] = &User{
		ID:        current.ID,
		Name:      name,
		Streaming: false,
		Owner:     false,
		Addr:      current.Addr,
		_write:    current.Write,
	}
	rooms.connected[current.ID] = room.ID
	room.notifyInfoChanged()
	usersJoinedTotal.Inc()

	v4, v6, err := rooms.config.TurnIPProvider.Get()
	if err != nil {
		return err
	}

	for _, user := range room.Users {
		if current.ID == user.ID || !user.Streaming {
			continue
		}
		if rooms.config.SFUMode {
			h, ok := room.SFUHosts[user.ID]
			if !ok {
				continue
			}
			if h.Track != nil {
				createViewerPC(room, rooms, user.ID, current.ID, v4, v6)
			} else {
				// Host PC exists but OnTrack hasn't fired yet; defer until the track arrives.
				h.Pending = append(h.Pending, current.ID)
			}
		} else {
			room.newSession(user.ID, current.ID, rooms, v4, v6)
		}
	}

	return nil
}
