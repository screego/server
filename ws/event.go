package ws

type Event interface {
	Execute(*Rooms, ClientInfo) error
}
