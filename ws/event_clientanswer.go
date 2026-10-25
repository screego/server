package ws

import (
	"encoding/json"
	"fmt"

	"github.com/pion/webrtc/v4"
	"github.com/rs/zerolog/log"
	"github.com/screego/server/ws/outgoing"
)

func init() {
	register("clientanswer", func() Event {
		return &ClientAnswer{}
	})
}

type ClientAnswer outgoing.P2PMessage

func (e *ClientAnswer) Execute(rooms *Rooms, current ClientInfo) error {
	room, err := rooms.CurrentRoom(current)
	if err != nil {
		return err
	}

	if rooms.config.SFUMode {
		var sd webrtc.SessionDescription
		if err := json.Unmarshal(e.Value, &sd); err != nil {
			return fmt.Errorf("SFU: unmarshal answer SDP: %w", err)
		}

		// Sharer answering the server's recvonly offer.
		if h := room.sfuHostBySession(e.SID); h != nil {
			if room.Users[current.ID] == nil || !room.Users[current.ID].Streaming {
				return fmt.Errorf("permission denied: not the sharing host")
			}
			if h.PC == nil {
				log.Debug().Msg("SFU: clientanswer for host but PC is nil")
				return nil
			}
			return h.PC.SetRemoteDescription(sd)
		}

		// Viewer answering the server's sendonly offer.
		session, ok := room.Sessions[e.SID]
		if !ok {
			log.Debug().Str("id", e.SID.String()).Msg("SFU: unknown session for clientanswer")
			return nil
		}
		if session.Client != current.ID {
			return fmt.Errorf("permission denied for session %s", e.SID)
		}
		if session.ViewerPC == nil {
			log.Debug().Str("id", e.SID.String()).Msg("SFU: clientanswer but ViewerPC is nil")
			return nil
		}
		return session.ViewerPC.SetRemoteDescription(sd)
	}

	session, ok := room.Sessions[e.SID]

	if !ok {
		log.Debug().Str("id", e.SID.String()).Msg("unknown session")
		return nil
	}

	if session.Client != current.ID {
		return fmt.Errorf("permission denied for session %s", e.SID)
	}

	room.Users[session.Host].WriteTimeout(outgoing.ClientAnswer(*e))

	return nil
}
