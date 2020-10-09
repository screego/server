package ws

import (
	"fmt"

	"github.com/screego/server/util"
)

func init() {
	register("join", func() Event {
		return &Join{}
	})
}

type Join struct {
	ID       string `json:"id"`
	UserName string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

func (e *Join) Execute(rooms *Rooms, current ClientInfo) error {
	if current.RoomID != "" {
		return fmt.Errorf("cannot join room, you are already in one")
	}

	room, ok := rooms.Rooms[e.ID]
	if !ok {
		return fmt.Errorf("room with id %s does not exist", e.ID)
	}
	name := e.UserName
	if current.Authenticated {
		name = current.AuthenticatedUser
	}
	if name == "" {
		name = util.NewName()
	}

	room.Users[current.ID] = &User{
		ID:        current.ID,
		Name:      name,
		Streaming: false,
		Owner:     false,
		Addr:      current.Addr,
		Write:     current.Write,
		Close:     current.Close,
	}
	room.notifyInfoChanged()

	for _, user := range room.Users {
		if current.ID == user.ID || !user.Streaming {
			continue
		}
		room.newSession(user.ID, current.ID, rooms)
	}

	return nil
}
