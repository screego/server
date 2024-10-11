package ws

func init() {
	register("name", func() Event {
		return &Name{}
	})
}

type Name struct {
	UserName string `json:"username"`
}

func (e *Name) Execute(rooms *Rooms, current ClientInfo) error {
	room, err := rooms.CurrentRoom(current)
	if err != nil {
		return err
	}

	room.Users[current.ID].Name = e.UserName

	room.notifyInfoChanged()
	return nil
}
