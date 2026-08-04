package ws

import (
	"fmt"

	"github.com/rs/xid"
)

func init() {
	register("subscribe", func() Event {
		return &Subscribe{}
	})
}

type Subscribe struct {
	SharerID xid.ID `json:"id"`
}

func (e *Subscribe) Execute(rooms *Rooms, current ClientInfo) error {
	room, err := rooms.CurrentRoom(current)
	if err != nil {
		return err
	}

	sharer, ok := room.Users[e.SharerID]
	if !ok || !sharer.Streaming {
		return fmt.Errorf("user %s is not streaming", e.SharerID)
	}

	v4, v6, err := rooms.config.TurnIPProvider.Get()
	if err != nil {
		return err
	}

	if rooms.config.SFUMode {
		h, ok := room.SFUHosts[e.SharerID]
		if !ok {
			return fmt.Errorf("no SFU host for sharer %s", e.SharerID)
		}
		if h.Track != nil {
			createViewerPC(room, rooms, e.SharerID, current.ID, v4, v6)
		} else {
			h.Pending = append(h.Pending, current.ID)
		}
	} else {
		room.newSession(e.SharerID, current.ID, rooms, v4, v6)
	}

	return nil
}
