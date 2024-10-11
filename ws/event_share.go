package ws

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

	for _, user := range room.Users {
		if current.ID == user.ID {
			continue
		}
		room.newSession(current.ID, user.ID, rooms, v4, v6)
	}

	room.notifyInfoChanged()
	return nil
}
