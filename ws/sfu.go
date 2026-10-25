package ws

import (
	"context"
	"encoding/json"
	"io"
	"net"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"
	"github.com/rs/xid"
	"github.com/rs/zerolog/log"
	"github.com/screego/server/ws/outgoing"
)

// Compile-time checks: SFU internal events implement the Event interface.
var _ Event = (*SFUHostTrack)(nil)
var _ Event = (*SFUIceCandidate)(nil)
var _ Event = (*SFUConnectionFailed)(nil)
var _ Event = (*SFUSendPLI)(nil)

// createHostPC creates a server-side recvonly PeerConnection for the sharing user.
// Sends hostsession + hostoffer to the sharer browser. Pion callbacks post internal SFU
// events into rooms.Incoming — they never touch Room state directly.
func createHostPC(room *Room, rooms *Rooms, sharerID xid.ID, v4addr, v6addr net.IP) {
	iceServers, pioncfg := buildICEServers(room, rooms, xid.New().String()+"host", v4addr, v6addr, room.Users[sharerID].Addr)
	sid := xid.New()

	h, ok := room.SFUHosts[sharerID]
	if !ok {
		return
	}
	h.SessionID = sid

	pc, err := rooms.webrtcAPI.NewPeerConnection(pioncfg)
	if err != nil {
		log.Error().Err(err).Msg("SFU: create host PC")
		return
	}
	h.PC = pc

	if _, err = pc.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo,
		webrtc.RTPTransceiverInit{Direction: webrtc.RTPTransceiverDirectionRecvonly}); err != nil {
		log.Error().Err(err).Msg("SFU: add recvonly transceiver")
		_ = pc.Close()
		h.PC = nil
		return
	}

	roomID := room.ID
	pc.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}
		rooms.Incoming <- ClientMessage{
			SkipConnectedCheck: true,
			Incoming: &SFUIceCandidate{
				RoomID: roomID, SessionID: sid,
				SharerID: sharerID, IsHost: true, Candidate: c,
			},
		}
	})

	pc.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		rooms.Incoming <- ClientMessage{
			SkipConnectedCheck: true,
			Incoming:           &SFUHostTrack{RoomID: roomID, SharerID: sharerID, Track: track, Receiver: receiver},
		}
	})

	pc.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		if state == webrtc.PeerConnectionStateFailed || state == webrtc.PeerConnectionStateDisconnected {
			rooms.Incoming <- ClientMessage{
				SkipConnectedCheck: true,
				Incoming:           &SFUConnectionFailed{RoomID: roomID, SharerID: sharerID, IsHost: true},
			}
		}
	})

	offer, err := pc.CreateOffer(nil)
	if err != nil {
		log.Error().Err(err).Msg("SFU: create host offer")
		_ = pc.Close()
		h.PC = nil
		return
	}
	if err = pc.SetLocalDescription(offer); err != nil {
		log.Error().Err(err).Msg("SFU: set host local description")
		_ = pc.Close()
		h.PC = nil
		return
	}

	sdpJSON, err := json.Marshal(offer)
	if err != nil {
		log.Error().Err(err).Msg("SFU: marshal host offer")
		_ = pc.Close()
		h.PC = nil
		return
	}

	sharer := room.Users[sharerID]
	sharer.WriteTimeout(outgoing.HostSession{ID: sid, Peer: sharerID, ICEServers: iceServers, Mode: "sfu"})
	sharer.WriteTimeout(outgoing.HostOffer{SID: sid, Value: sdpJSON})
}

// createViewerPC creates a server-side sendonly PeerConnection for a viewer and sends
// clientsession + hostoffer to the viewer browser.
func createViewerPC(room *Room, rooms *Rooms, sharerID, viewerID xid.ID, v4addr, v6addr net.IP) {
	iceServers, pioncfg := buildICEServers(room, rooms, xid.New().String()+"client", v4addr, v6addr, room.Users[viewerID].Addr)
	sid := xid.New()

	pc, err := rooms.webrtcAPI.NewPeerConnection(pioncfg)
	if err != nil {
		log.Error().Err(err).Msg("SFU: create viewer PC")
		return
	}

	roomID := room.ID

	h := room.SFUHosts[sharerID]
	if h != nil && h.Track != nil {
		sender, err := pc.AddTrack(h.Track)
		if err != nil {
			log.Error().Err(err).Msg("SFU: add shared track to viewer PC")
			_ = pc.Close()
			return
		}
		// Forward PLI from the viewer browser back to the sharer so it re-sends a keyframe.
		go func() {
			rtcpBuf := make([]byte, 1500)
			for {
				n, _, err := sender.Read(rtcpBuf)
				if err != nil {
					return
				}
				pkts, err := rtcp.Unmarshal(rtcpBuf[:n])
				if err != nil {
					continue
				}
				for _, pkt := range pkts {
					if _, ok := pkt.(*rtcp.PictureLossIndication); ok {
						rooms.Incoming <- ClientMessage{
							SkipConnectedCheck: true,
							Incoming:           &SFUSendPLI{RoomID: roomID, SharerID: sharerID},
						}
					}
				}
			}
		}()
	}

	room.Sessions[sid] = &RoomSession{Host: sharerID, Client: viewerID, ViewerPC: pc}
	sessionCreatedTotal.Inc()

	pc.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}
		rooms.Incoming <- ClientMessage{
			SkipConnectedCheck: true,
			Incoming: &SFUIceCandidate{
				RoomID: roomID, SessionID: sid,
				SharerID: sharerID, IsHost: false, Candidate: c,
			},
		}
	})

	pc.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		if state == webrtc.PeerConnectionStateConnected {
			// New viewer connected — ask sharer for a keyframe so the viewer can start decoding.
			rooms.Incoming <- ClientMessage{
				SkipConnectedCheck: true,
				Incoming:           &SFUSendPLI{RoomID: roomID, SharerID: sharerID},
			}
		}
		if state == webrtc.PeerConnectionStateFailed || state == webrtc.PeerConnectionStateDisconnected {
			rooms.Incoming <- ClientMessage{
				SkipConnectedCheck: true,
				Incoming:           &SFUConnectionFailed{RoomID: roomID, SharerID: sharerID, IsHost: false, SessionID: sid},
			}
		}
	})

	offer, err := pc.CreateOffer(nil)
	if err != nil {
		log.Error().Err(err).Msg("SFU: create viewer offer")
		_ = pc.Close()
		delete(room.Sessions, sid)
		return
	}
	if err = pc.SetLocalDescription(offer); err != nil {
		log.Error().Err(err).Msg("SFU: set viewer local description")
		_ = pc.Close()
		delete(room.Sessions, sid)
		return
	}

	sdpJSON, err := json.Marshal(offer)
	if err != nil {
		log.Error().Err(err).Msg("SFU: marshal viewer offer")
		_ = pc.Close()
		delete(room.Sessions, sid)
		return
	}

	viewer := room.Users[viewerID]
	viewer.WriteTimeout(outgoing.ClientSession{ID: sid, Peer: sharerID, ICEServers: iceServers})
	viewer.WriteTimeout(outgoing.HostOffer{SID: sid, Value: sdpJSON})
}

// startRTPForward pumps RTP from the sharer's track to all viewers via the shared track
// and drains RTCP from the receiver so pion's internal buffers don't stall.
// PLI forwarding (viewer→sharer) is handled separately in createViewerPC via sender.Read().
func startRTPForward(ctx context.Context, track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver,
	shared *webrtc.TrackLocalStaticRTP) {

	go func() {
		buf := make([]byte, 1500)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			n, _, err := track.Read(buf)
			if err != nil {
				return
			}
			if _, err = shared.Write(buf[:n]); err != nil && err != io.ErrClosedPipe {
				return
			}
		}
	}()

	// Drain RTCP from the sharer's receiver so pion's internal buffers don't stall.
	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, err := receiver.Read(rtcpBuf); err != nil {
				return
			}
		}
	}()
}

// buildICEServers returns:
//   - outgoing ICE servers to send to the browser (STUN or TURN with credentials)
//   - a pion Configuration for the server-side PeerConnection
//
// Pion is the server endpoint — it never needs TURN to reach itself.
// In STUN mode pion gets the same STUN URL so it can discover its reflexive address.
// In TURN mode pion gets only STUN (derived from the TURN address) so it discovers
// its public IP without allocating a relay it will never use.
func buildICEServers(room *Room, rooms *Rooms, credKey string, v4addr, v6addr, clientAddr net.IP) ([]outgoing.ICEServer, webrtc.Configuration) {
	var out []outgoing.ICEServer
	var pionServers []webrtc.ICEServer

	switch room.Mode {
	case ConnectionSTUN:
		urls := rooms.addresses("stun", v4addr, v6addr, false)
		out = []outgoing.ICEServer{{URLs: urls}}
		pionServers = []webrtc.ICEServer{{URLs: urls}}
	case ConnectionTURN:
		name, pw := rooms.turnServer.Credentials(credKey, clientAddr)
		turnURLs := rooms.addresses("turn", v4addr, v6addr, true)
		out = []outgoing.ICEServer{{URLs: turnURLs, Username: name, Credential: pw}}
		// Pion only needs STUN to discover its reflexive address; not TURN.
		stunURLs := rooms.addresses("stun", v4addr, v6addr, false)
		if len(stunURLs) > 0 {
			pionServers = []webrtc.ICEServer{{URLs: stunURLs}}
		}
	}

	return out, webrtc.Configuration{ICEServers: pionServers}
}
