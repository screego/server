package ws

import (
	"fmt"
)

func init() {
	register("share", func() Event {
		return &StartShare{}
	})
}

type StartShare struct {
}

func (e *StartShare) Execute(rooms *Rooms, current ClientInfo) error {
	if current.RoomID == "" {
		return fmt.Errorf("not in a room")
	}

	room, ok := rooms.Rooms[current.RoomID]
	if !ok {
		return fmt.Errorf("room with id %s does not exist", current.RoomID)
	}

	room.Users[current.ID].Streaming = true

	for _, user := range room.Users {
		if current.ID == user.ID {
			continue
		}
		room.newSession(current.ID, user.ID, rooms)
	}

	room.notifyInfoChanged()
	return nil
}
