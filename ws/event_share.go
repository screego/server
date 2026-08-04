package ws

import "github.com/rs/xid"

func init() {
	register("share", func() Event {
		return &StartShare{}
	})
}

type StartShare struct{}

func (e *StartShare) Execute(rooms *Rooms, current ClientInfo) error {
	room, err := rooms.CurrentRoom(current)
	if err != nil {
		return err
	}

	room.Users[current.ID].Streaming = true

	v4, v6, err := rooms.config.TurnIPProvider.Get()
	if err != nil {
		return err
	}

	if rooms.config.SFUMode {
		if room.SFUHosts == nil {
			room.SFUHosts = make(map[xid.ID]*SFUHost)
		}
		// Collect all other users as pending viewers for this sharer.
		// SFUHostTrack.Execute will create their ViewerPCs once the track arrives.
		pending := make([]xid.ID, 0, len(room.Users)-1)
		for _, user := range room.Users {
			if user.ID != current.ID {
				pending = append(pending, user.ID)
			}
		}
		room.SFUHosts[current.ID] = &SFUHost{Pending: pending}
		createHostPC(room, rooms, current.ID, v4, v6)
	} else {
		for _, user := range room.Users {
			if current.ID == user.ID {
				continue
			}
			room.newSession(current.ID, user.ID, rooms, v4, v6)
		}
	}

	room.notifyInfoChanged()
	return nil
}
