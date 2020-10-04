package ws

import (
	"fmt"

	"github.com/screego/server/ws/outgoing"
)

func init() {
	register("clientanswer", func() Event {
		return &ClientAnswer{}
	})
}

type ClientAnswer outgoing.P2PMessage

func (e *ClientAnswer) Execute(rooms *Rooms, current ClientInfo) error {
	if current.RoomID == "" {
		return fmt.Errorf("not in a room")
	}

	room, ok := rooms.Rooms[current.RoomID]
	if !ok {
		return fmt.Errorf("room with id %s does not exist", current.RoomID)
	}

	session, ok := room.Sessions[e.SID]

	if !ok || session.Client != current.ID {
		return fmt.Errorf("session with id %s does not exist", current.RoomID)
	}

	room.Users[session.Host].Write <- outgoing.ClientAnswer(*e)

	return nil
}
