package ws

type Health struct {
	Response chan int
}

func (e *Health) Execute(rooms *Rooms, current ClientInfo) error {
	writeTimeout(e.Response, len(rooms.connected))
	return nil
}
