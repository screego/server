package ws

import (
	"fmt"
)

func init() {
	register("name", func() Event {
		return &Name{}
	})
}

type Name struct {
	UserName string `json:"username"`
}

func (e *Name) Execute(rooms *Rooms, current ClientInfo) error {
	if current.RoomID == "" {
		return fmt.Errorf("not in a room")
	}

	room, ok := rooms.Rooms[current.RoomID]
	if !ok {
		return fmt.Errorf("room with id %s does not exist", current.RoomID)
	}

	room.Users[current.ID].Name = e.UserName

	room.notifyInfoChanged()
	return nil
}
