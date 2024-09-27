package ws

type Connected struct{}

func (e Connected) Execute(rooms *Rooms, current ClientInfo) error {
	rooms.connected[current.ID] = true
	return nil
}
