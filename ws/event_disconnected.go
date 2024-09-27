package ws

import (
	"bytes"

	"github.com/gorilla/websocket"
	"github.com/screego/server/ws/outgoing"
)

type Disconnected struct {
	Code   int
	Reason string
}

func (e *Disconnected) Execute(rooms *Rooms, current ClientInfo) error {
	e.executeNoError(rooms, current)
	return nil
}

func (e *Disconnected) executeNoError(rooms *Rooms, current ClientInfo) {
	delete(rooms.connected, current.ID)
	current.Write <- outgoing.CloseWriter{Code: e.Code, Reason: e.Reason}

	if current.RoomID == "" {
		return
	}

	room, ok := rooms.Rooms[current.RoomID]
	if !ok {
		// room may already be removed
		return
	}

	user, ok := room.Users[current.ID]

	if !ok {
		// room may already be removed
		return
	}

	delete(room.Users, current.ID)
	usersLeftTotal.Inc()

	for id, session := range room.Sessions {
		if bytes.Equal(session.Client.Bytes(), current.ID.Bytes()) {
			host, ok := room.Users[session.Host]
			if ok {
				host.Write <- outgoing.EndShare(id)
			}
			room.closeSession(rooms, id)
		}
		if bytes.Equal(session.Host.Bytes(), current.ID.Bytes()) {
			client, ok := room.Users[session.Client]
			if ok {
				client.Write <- outgoing.EndShare(id)
			}
			room.closeSession(rooms, id)
		}
	}

	if user.Owner && room.CloseOnOwnerLeave {
		for _, member := range room.Users {
			delete(rooms.connected, member.ID)
			member.Write <- outgoing.CloseWriter{Code: websocket.CloseNormalClosure, Reason: CloseOwnerLeft}
		}
		rooms.closeRoom(current.RoomID)
		return
	}

	if len(room.Users) == 0 {
		rooms.closeRoom(current.RoomID)
		return
	}

	room.notifyInfoChanged()
}
