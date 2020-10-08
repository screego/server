package ws

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

	if user.Owner && room.CloseOnOwnerLeave {
		for _, member := range room.Users {
			member.Close <- CloseOwnerLeft
		}
		delete(rooms.Rooms, current.RoomID)
		return nil
	}

	if len(room.Users) == 0 {
		delete(rooms.Rooms, current.RoomID)
		return nil
	}

	room.notifyInfoChanged()

	return nil
}
