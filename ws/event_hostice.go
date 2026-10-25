package ws

import (
	"encoding/json"
	"fmt"

	"github.com/pion/webrtc/v4"
	"github.com/rs/zerolog/log"
	"github.com/screego/server/ws/outgoing"
)

func init() {
	register("hostice", func() Event {
		return &HostICE{}
	})
}

type HostICE outgoing.P2PMessage

func (e *HostICE) Execute(rooms *Rooms, current ClientInfo) error {
	room, err := rooms.CurrentRoom(current)
	if err != nil {
		return err
	}

	if rooms.config.SFUMode {
		// Sharer sends ICE for its upload connection to the server.
		h := room.sfuHostBySession(e.SID)
		if h == nil {
			log.Debug().Str("id", e.SID.String()).Msg("SFU: unexpected hostice for unknown session")
			return nil
		}
		if room.Users[current.ID] == nil || !room.Users[current.ID].Streaming {
			return fmt.Errorf("permission denied: not the sharing host")
		}
		if h.PC == nil {
			return nil
		}
		var init webrtc.ICECandidateInit
		if err := json.Unmarshal(e.Value, &init); err != nil {
			return fmt.Errorf("SFU: unmarshal host ICE candidate: %w", err)
		}
		return h.PC.AddICECandidate(init)
	}

	session, ok := room.Sessions[e.SID]

	if !ok {
		log.Debug().Str("id", e.SID.String()).Msg("unknown session")
		return nil
	}

	if session.Host != current.ID {
		return fmt.Errorf("permission denied for session %s", e.SID)
	}

	room.Users[session.Client].WriteTimeout(outgoing.HostICE(*e))

	return nil
}
