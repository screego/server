package ws

import (
	"bytes"
	"github.com/screego/server/ws/outgoing"
)

type Disconnected struct {
}

func (e *Disconnected) Execute(rooms *Rooms, current ClientInfo) error {
	if current.RoomID == "" {
		return nil
	}

	room, ok := rooms.Rooms[current.RoomID]
	if !ok {
		// room may already be removed
		return nil
	}

	user, ok := room.Users[current.ID]

	if !ok {
		// room may already be removed
		return nil
	}

	current.Close <- CloseDone
	delete(room.Users, current.ID)
	usersLeftTotal.Inc()

	for id, session := range room.Sessions {
		if bytes.Equal(session.Client.Bytes(), current.ID.Bytes()) {
			host, ok := room.Users[session.Host]
			if ok {
				host.Write <- outgoing.EndShare(id)
			}
			room.closeSession(id)
		}
		if bytes.Equal(session.Host.Bytes(), current.ID.Bytes()) {
			client, ok := room.Users[session.Client]
			if ok {
				client.Write <- outgoing.EndShare(id)
			}
			room.closeSession(id)
		}
	}

	if user.Owner && room.CloseOnOwnerLeave {
		for _, member := range room.Users {
			member.Close <- CloseOwnerLeft
		}
		rooms.closeRoom(current.RoomID)
		return nil
	}

	if len(room.Users) == 0 {
		rooms.closeRoom(current.RoomID)
		return nil
	}

	room.notifyInfoChanged()

	return nil
}
