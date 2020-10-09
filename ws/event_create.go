package ws

import (
	"errors"
	"fmt"

	"github.com/rs/xid"
	"github.com/screego/server/config"
	"github.com/screego/server/util"
)

func init() {
	register("create", func() Event {
		return &Create{}
	})
}

type Create struct {
	ID                string         `json:"id"`
	Mode              ConnectionMode `json:"mode"`
	CloseOnOwnerLeave bool           `json:"closeOnOwnerLeave"`
	UserName          string         `json:"username"`
}

func (e *Create) Execute(rooms *Rooms, current ClientInfo) error {
	if current.RoomID != "" {
		return fmt.Errorf("cannot join room, you are already in one")
	}

	if _, ok := rooms.Rooms[e.ID]; ok {
		return fmt.Errorf("room with id %s does already existn", e.ID)
	}

	name := e.UserName
	if current.Authenticated {
		name = current.AuthenticatedUser
	}
	if name == "" {
		name = util.NewName()
	}

	switch rooms.config.AuthMode {
	case config.AuthModeNone:
	case config.AuthModeAll:
		if !current.Authenticated {
			return errors.New("you need to login")
		}
	case config.AuthModeTurn:
		if e.Mode != ConnectionSTUN && e.Mode != ConnectionLocal && !current.Authenticated {
			return errors.New("you need to login")
		}
	default:
		return errors.New("invalid authmode:" + rooms.config.AuthMode)
	}

	room := &Room{
		ID:                e.ID,
		CloseOnOwnerLeave: e.CloseOnOwnerLeave,
		Mode:              e.Mode,
		Sessions:          map[xid.ID]*RoomSession{},
		Users: map[xid.ID]*User{
			current.ID: {
				ID:        current.ID,
				Name:      name,
				Streaming: false,
				Owner:     true,
				Addr:      current.Addr,
				Write:     current.Write,
				Close:     current.Close,
			},
		},
	}
	rooms.Rooms[e.ID] = room
	room.notifyInfoChanged()
	return nil
}
