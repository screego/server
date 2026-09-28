package ws

import (
	"github.com/rs/xid"
	"github.com/screego/server/ws/outgoing"
)

func init() {
	register("unsubscribe", func() Event {
		return &Unsubscribe{}
	})
}

type Unsubscribe struct {
	SharerID xid.ID `json:"id"`
}

func (e *Unsubscribe) Execute(rooms *Rooms, current ClientInfo) error {
	room, err := rooms.CurrentRoom(current)
	if err != nil {
		return err
	}

	if rooms.config.SFUMode {
		// Remove from Pending if the track hadn't arrived yet.
		if h, ok := room.SFUHosts[e.SharerID]; ok {
			filtered := h.Pending[:0]
			for _, id := range h.Pending {
				if id != current.ID {
					filtered = append(filtered, id)
				}
			}
			h.Pending = filtered
		}
	}

	sid, session, ok := room.sessionByPeers(e.SharerID, current.ID)
	if !ok {
		// Still pending (track not arrived yet); removal above is sufficient.
		return nil
	}

	if session.ViewerPC != nil && rooms.config.SFUMode {
		_ = session.ViewerPC.Close()
	} else if !rooms.config.SFUMode {
		// P2P: notify the sharer so it closes its host PC for this viewer.
		if sharer, ok := room.Users[e.SharerID]; ok {
			sharer.WriteTimeout(outgoing.EndShare(sid))
		}
	}

	room.closeSession(rooms, sid)
	return nil
}
