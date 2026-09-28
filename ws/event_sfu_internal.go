package ws

import (
	"context"
	"encoding/json"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"
	"github.com/rs/xid"
	"github.com/rs/zerolog/log"
	"github.com/screego/server/ws/outgoing"
)

// SFUHostTrack is posted by the pion OnTrack callback when a sharer's video arrives.
// Running in the main event loop, it is safe to mutate Room state.
type SFUHostTrack struct {
	RoomID   string
	SharerID xid.ID // user ID of the sharer
	Track    *webrtc.TrackRemote
	Receiver *webrtc.RTPReceiver
}

func (e *SFUHostTrack) Execute(rooms *Rooms, _ ClientInfo) error {
	room, ok := rooms.Rooms[e.RoomID]
	if !ok {
		return nil
	}

	h, ok := room.SFUHosts[e.SharerID]
	if !ok {
		return nil
	}

	shared, err := webrtc.NewTrackLocalStaticRTP(
		e.Track.Codec().RTPCodecCapability, "video", "screego-sfu")
	if err != nil {
		log.Error().Err(err).Msg("SFU: create shared track")
		return nil
	}
	h.Track = shared
	h.SSRC = uint32(e.Track.SSRC())

	ctx, cancel := context.WithCancel(context.Background())
	h.Cancel = cancel

	startRTPForward(ctx, e.Track, e.Receiver, shared)

	// Connect any viewers who joined before the track arrived.
	pending := h.Pending
	h.Pending = nil

	v4addr, v6addr, err := rooms.config.TurnIPProvider.Get()
	if err != nil {
		log.Error().Err(err).Msg("SFU: get TURN IPs for pending viewers")
		return nil
	}

	for _, viewerID := range pending {
		if _, ok := room.Users[viewerID]; !ok {
			continue
		}
		createViewerPC(room, rooms, e.SharerID, viewerID, v4addr, v6addr)
	}

	return nil
}

// SFUIceCandidate is posted by pion OnICECandidate callbacks and routes the server's
// ICE candidates to the correct browser peer.
//
// ICE routing (naming inherited from P2P era):
//   - server → sharer browser : outgoing.ClientICE  (browser routes via host.current[sid])
//   - server → viewer browser : outgoing.HostICE    (browser routes via client.current[sid])
type SFUIceCandidate struct {
	RoomID    string
	SessionID xid.ID
	SharerID  xid.ID // user ID of the sharer (populated for both host and viewer candidates)
	IsHost    bool
	Candidate *webrtc.ICECandidate
}

func (e *SFUIceCandidate) Execute(rooms *Rooms, _ ClientInfo) error {
	room, ok := rooms.Rooms[e.RoomID]
	if !ok {
		return nil
	}

	raw, err := json.Marshal(e.Candidate.ToJSON())
	if err != nil {
		log.Error().Err(err).Msg("SFU: marshal ICE candidate")
		return nil
	}

	if e.IsHost {
		// Send to the sharer browser via clientice (routes to host.current[sid] in browser).
		if sharer, ok := room.Users[e.SharerID]; ok {
			sharer.WriteTimeout(outgoing.ClientICE{SID: e.SessionID, Value: raw})
		}
	} else {
		session, ok := room.Sessions[e.SessionID]
		if !ok {
			return nil
		}
		viewer, ok := room.Users[session.Client]
		if !ok {
			return nil
		}
		// hostice → browser routes to client.current[sid]
		viewer.WriteTimeout(outgoing.HostICE{SID: e.SessionID, Value: raw})
	}
	return nil
}

// SFUSendPLI is posted by viewer PC callbacks to request a keyframe from the sharer.
// Safe to call from any goroutine — always executed inside the main event loop.
type SFUSendPLI struct {
	RoomID   string
	SharerID xid.ID
}

func (e *SFUSendPLI) Execute(rooms *Rooms, _ ClientInfo) error {
	room, ok := rooms.Rooms[e.RoomID]
	if !ok {
		return nil
	}
	h, ok := room.SFUHosts[e.SharerID]
	if !ok || h.PC == nil || h.SSRC == 0 {
		return nil
	}
	return h.PC.WriteRTCP([]rtcp.Packet{
		&rtcp.PictureLossIndication{MediaSSRC: h.SSRC},
	})
}

// SFUConnectionFailed is posted when a pion PC enters failed/disconnected state.
type SFUConnectionFailed struct {
	RoomID    string
	SharerID  xid.ID // user ID of the sharer (always set)
	IsHost    bool
	SessionID xid.ID // only set when IsHost=false
}

func (e *SFUConnectionFailed) Execute(rooms *Rooms, _ ClientInfo) error {
	room, ok := rooms.Rooms[e.RoomID]
	if !ok {
		return nil
	}

	if e.IsHost {
		room.closeSFUHost(e.SharerID)
		// Close all viewer sessions for this sharer.
		for id, session := range room.Sessions {
			if session.Host != e.SharerID {
				continue
			}
			if viewer, ok := room.Users[session.Client]; ok {
				viewer.WriteTimeout(outgoing.EndShare(id))
			}
			room.closeSession(rooms, id)
		}
		if sharer, ok := room.Users[e.SharerID]; ok {
			sharer.Streaming = false
		}
		room.notifyInfoChanged()
		return nil
	}

	// A single viewer's PC failed.
	session, ok := room.Sessions[e.SessionID]
	if !ok {
		return nil
	}
	if viewer, ok := room.Users[session.Client]; ok {
		viewer.WriteTimeout(outgoing.EndShare(e.SessionID))
	}
	room.closeSession(rooms, e.SessionID)
	return nil
}
