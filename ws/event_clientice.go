package ws

import (
	"encoding/json"
	"fmt"

	"github.com/pion/webrtc/v4"
	"github.com/rs/zerolog/log"
	"github.com/screego/server/ws/outgoing"
)

func init() {
	register("clientice", func() Event {
		return &ClientICE{}
	})
}

type ClientICE outgoing.P2PMessage

func (e *ClientICE) Execute(rooms *Rooms, current ClientInfo) error {
	room, err := rooms.CurrentRoom(current)
	if err != nil {
		return err
	}

	if rooms.config.SFUMode {
		// Viewer sends ICE for its download connection to the server's ViewerPC.
		session, ok := room.Sessions[e.SID]
		if !ok {
			log.Debug().Str("id", e.SID.String()).Msg("SFU: unknown session for clientice")
			return nil
		}
		if session.Client != current.ID {
			return fmt.Errorf("permission denied for session %s", e.SID)
		}
		if session.ViewerPC == nil {
			return nil
		}
		var init webrtc.ICECandidateInit
		if err := json.Unmarshal(e.Value, &init); err != nil {
			return fmt.Errorf("SFU: unmarshal viewer ICE candidate: %w", err)
		}
		return session.ViewerPC.AddICECandidate(init)
	}

	session, ok := room.Sessions[e.SID]

	if !ok {
		log.Debug().Str("id", e.SID.String()).Msg("unknown session")
		return nil
	}

	if session.Client != current.ID {
		return fmt.Errorf("permission denied for session %s", e.SID)
	}

	room.Users[session.Host].WriteTimeout(outgoing.ClientICE(*e))

	return nil
}
