package ws

import (
	"bytes"
	"fmt"

	"github.com/screego/server/ws/outgoing"
)

func init() {
	register("stopshare", func() Event {
		return &StopShare{}
	})
}

type StopShare struct {
}

func (e *StopShare) Execute(rooms *Rooms, current ClientInfo) error {
	if current.RoomID == "" {
		return fmt.Errorf("not in a room")
	}

	room, ok := rooms.Rooms[current.RoomID]
	if !ok {
		return fmt.Errorf("room with id %s does not exist", current.RoomID)
	}

	room.Users[current.ID].Streaming = false
	for id, session := range room.Sessions {
		if bytes.Equal(session.Host.Bytes(), current.ID.Bytes()) {
			client, ok := room.Users[session.Client]
			if ok {
				client.Write <- outgoing.EndShare(id)
			}
			delete(room.Sessions, id)
		}
	}

	room.notifyInfoChanged()
	return nil
}
