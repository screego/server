package ws

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/screego/server/ws/outgoing"
)

func init() {
	register("hostoffer", func() Event {
		return &HostOffer{}
	})
}

type HostOffer outgoing.P2PMessage

func (e *HostOffer) Execute(rooms *Rooms, current ClientInfo) error {
	room, err := rooms.CurrentRoom(current)
	if err != nil {
		return err
	}

	session, ok := room.Sessions[e.SID]

	if !ok {
		log.Debug().Str("id", e.SID.String()).Msg("unknown session")
		return nil
	}

	if session.Host != current.ID {
		return fmt.Errorf("permission denied for session %s", e.SID)
	}

	room.Users[session.Client].Write <- outgoing.HostOffer(*e)

	return nil
}
