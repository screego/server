package ws

import (
	"context"
	"net"
	"testing"

	"github.com/rs/xid"
	"github.com/screego/server/config"
	"github.com/screego/server/config/ipdns"
	"github.com/screego/server/ws/outgoing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- test helpers ---

type mockTurnServer struct {
	disallowed []string
}

func (m *mockTurnServer) Credentials(id string, addr net.IP) (string, string) {
	return "testuser-" + id, "testpass"
}
func (m *mockTurnServer) Disallow(username string) {
	m.disallowed = append(m.disallowed, username)
}

type staticIPProvider struct{ v4, v6 net.IP }

func (s staticIPProvider) Get() (net.IP, net.IP, error) { return s.v4, s.v6, nil }

var _ ipdns.Provider = staticIPProvider{}

func makeSFURooms() *Rooms {
	ts := &mockTurnServer{}
	return &Rooms{
		Rooms:      map[string]*Room{},
		Incoming:   make(chan ClientMessage, 20),
		connected:  map[xid.ID]string{},
		turnServer: ts,
		config: config.Config{
			SFUMode:  true,
			TurnPort: "3478",
			TurnIPProvider: staticIPProvider{
				v4: net.ParseIP("127.0.0.1"),
			},
		},
	}
}

func makeTestRoom(id string, mode ConnectionMode) *Room {
	return &Room{
		ID:       id,
		Mode:     mode,
		Users:    map[xid.ID]*User{},
		Sessions: map[xid.ID]*RoomSession{},
	}
}

func makeTestUser(streaming bool) (*User, chan outgoing.Message) {
	ch := make(chan outgoing.Message, 10)
	u := &User{
		ID:        xid.New(),
		Streaming: streaming,
		_write:    ch,
	}
	return u, ch
}

// --- closeSFUHost ---

func TestCloseSFUHost_NilSafe(t *testing.T) {
	room := makeTestRoom("r", ConnectionLocal)
	// No SFUHosts at all — must not panic.
	assert.NotPanics(t, func() { room.closeSFUHost(xid.New()) })
	assert.NotPanics(t, room.closeAllSFUHosts)
	assert.Empty(t, room.SFUHosts)
}

func TestCloseSFUHost_CallsCancel(t *testing.T) {
	room := makeTestRoom("r", ConnectionLocal)
	sharerID := xid.New()
	cancelled := false
	ctx, cancel := context.WithCancel(context.Background())
	room.SFUHosts = map[xid.ID]*SFUHost{
		sharerID: {
			Cancel: func() {
				cancelled = true
				cancel()
			},
		},
	}
	room.closeSFUHost(sharerID)
	assert.True(t, cancelled)
	assert.NotContains(t, room.SFUHosts, sharerID)
	assert.Equal(t, context.Canceled, ctx.Err())
}

func TestCloseSFUHost_ClearsPending(t *testing.T) {
	room := makeTestRoom("r", ConnectionLocal)
	sharerID := xid.New()
	room.SFUHosts = map[xid.ID]*SFUHost{
		sharerID: {Pending: []xid.ID{xid.New(), xid.New()}},
	}
	room.closeSFUHost(sharerID)
	assert.NotContains(t, room.SFUHosts, sharerID)
}

// --- SFUConnectionFailed ---

func TestSFUConnectionFailed_UnknownRoom(t *testing.T) {
	rooms := makeSFURooms()
	ev := &SFUConnectionFailed{RoomID: "no-such-room", IsHost: true}
	// Must not panic.
	require.NoError(t, ev.Execute(rooms, ClientInfo{}))
}

func TestSFUConnectionFailed_Host_MarksUsersNotStreaming(t *testing.T) {
	rooms := makeSFURooms()
	room := makeTestRoom("r", ConnectionLocal)
	rooms.Rooms["r"] = room

	host, _ := makeTestUser(true)
	viewer, viewerCh := makeTestUser(false)
	room.Users[host.ID] = host
	room.Users[viewer.ID] = viewer

	sid := xid.New()
	room.Sessions[sid] = &RoomSession{Host: host.ID, Client: viewer.ID}
	room.SFUHosts = map[xid.ID]*SFUHost{host.ID: {}}

	ev := &SFUConnectionFailed{RoomID: "r", SharerID: host.ID, IsHost: true}
	require.NoError(t, ev.Execute(rooms, ClientInfo{}))

	assert.False(t, host.Streaming, "host should be marked not streaming")
	assert.Empty(t, room.Sessions, "all sessions should be removed")
	assert.NotContains(t, room.SFUHosts, host.ID)

	var msg outgoing.Message
	select {
	case msg = <-viewerCh:
	default:
		t.Fatal("viewer should have received endshare")
	}
	_, ok := msg.(outgoing.EndShare)
	assert.True(t, ok, "viewer should receive EndShare message")
}

func TestSFUConnectionFailed_Viewer_ClosesOneSession(t *testing.T) {
	rooms := makeSFURooms()
	room := makeTestRoom("r", ConnectionLocal)
	rooms.Rooms["r"] = room

	host, _ := makeTestUser(true)
	viewer1, viewer1Ch := makeTestUser(false)
	viewer2, viewer2Ch := makeTestUser(false)
	room.Users[host.ID] = host
	room.Users[viewer1.ID] = viewer1
	room.Users[viewer2.ID] = viewer2

	sid1 := xid.New()
	sid2 := xid.New()
	room.Sessions[sid1] = &RoomSession{Host: host.ID, Client: viewer1.ID}
	room.Sessions[sid2] = &RoomSession{Host: host.ID, Client: viewer2.ID}

	ev := &SFUConnectionFailed{RoomID: "r", SharerID: host.ID, IsHost: false, SessionID: sid1}
	require.NoError(t, ev.Execute(rooms, ClientInfo{}))

	assert.NotContains(t, room.Sessions, sid1, "sid1 should be removed")
	assert.Contains(t, room.Sessions, sid2, "sid2 should remain")

	select {
	case <-viewer1Ch:
		// expected endshare
	default:
		t.Fatal("viewer1 should receive endshare")
	}

	select {
	case <-viewer2Ch:
		t.Fatal("viewer2 should not receive anything")
	default:
		// expected
	}
}

func TestSFUConnectionFailed_Viewer_UnknownSession(t *testing.T) {
	rooms := makeSFURooms()
	room := makeTestRoom("r", ConnectionLocal)
	rooms.Rooms["r"] = room

	ev := &SFUConnectionFailed{RoomID: "r", IsHost: false, SessionID: xid.New()}
	// Must not panic on unknown session.
	require.NoError(t, ev.Execute(rooms, ClientInfo{}))
}

// --- buildICEServers ---

func TestBuildICEServers_Local(t *testing.T) {
	rooms := makeSFURooms()
	room := makeTestRoom("r", ConnectionLocal)
	v4 := net.ParseIP("1.2.3.4")

	out, cfg := buildICEServers(room, rooms, "k", v4, nil, nil)
	assert.Empty(t, out)
	assert.Empty(t, cfg.ICEServers)
}

func TestBuildICEServers_STUN(t *testing.T) {
	rooms := makeSFURooms()
	room := makeTestRoom("r", ConnectionSTUN)
	v4 := net.ParseIP("1.2.3.4")

	out, cfg := buildICEServers(room, rooms, "k", v4, nil, nil)
	require.Len(t, out, 1)
	require.Len(t, cfg.ICEServers, 1)
	assert.Equal(t, out[0].URLs, cfg.ICEServers[0].URLs)
	assert.Contains(t, out[0].URLs[0], "stun:")
	assert.Contains(t, out[0].URLs[0], "1.2.3.4")
	assert.Contains(t, out[0].URLs[0], "3478")
}

func TestBuildICEServers_TURN_HasCredentials(t *testing.T) {
	rooms := makeSFURooms()
	room := makeTestRoom("r", ConnectionTURN)
	v4 := net.ParseIP("1.2.3.4")
	clientAddr := net.ParseIP("9.9.9.9")

	out, cfg := buildICEServers(room, rooms, "mykey", v4, nil, clientAddr)

	// Browser gets TURN credentials.
	require.Len(t, out, 1)
	assert.NotEmpty(t, out[0].Username)
	assert.NotEmpty(t, out[0].Credential)
	assert.Contains(t, out[0].URLs[0], "turn:")

	// Pion gets STUN only — it is the server endpoint, never needs its own relay.
	require.Len(t, cfg.ICEServers, 1)
	assert.Contains(t, cfg.ICEServers[0].URLs[0], "stun:")
	assert.Empty(t, cfg.ICEServers[0].Username, "pion should not get TURN credentials")
}

func TestBuildICEServers_STUN_WithV6(t *testing.T) {
	rooms := makeSFURooms()
	room := makeTestRoom("r", ConnectionSTUN)
	v4 := net.ParseIP("1.2.3.4")
	v6 := net.ParseIP("2001:db8::1")

	out, _ := buildICEServers(room, rooms, "k", v4, v6, nil)
	require.Len(t, out, 1)
	// Both v4 and v6 URLs in the same server entry.
	assert.Len(t, out[0].URLs, 2)
}

// --- SFUMode=false leaves P2P path unchanged ---

func TestShareEvent_SFUMode_QueuesExistingViewers(t *testing.T) {
	rooms := makeSFURooms()
	rooms.webrtcAPI = newWebRTCAPI()
	room := makeTestRoom("r", ConnectionLocal)
	rooms.Rooms["r"] = room

	host, _ := makeTestUser(false)
	viewer1, _ := makeTestUser(false)
	viewer2, _ := makeTestUser(false)
	room.Users[host.ID] = host
	room.Users[viewer1.ID] = viewer1
	room.Users[viewer2.ID] = viewer2
	rooms.connected[host.ID] = "r"
	rooms.connected[viewer1.ID] = "r"
	rooms.connected[viewer2.ID] = "r"

	ev := &StartShare{}
	require.NoError(t, ev.Execute(rooms, ClientInfo{ID: host.ID}))

	// Both viewers should be queued in the sharer's SFUHost entry.
	require.Contains(t, room.SFUHosts, host.ID)
	pending := room.SFUHosts[host.ID].Pending
	assert.Len(t, pending, 2)
	assert.Contains(t, pending, viewer1.ID)
	assert.Contains(t, pending, viewer2.ID)
	assert.NotContains(t, pending, host.ID)
}

func TestShareEvent_P2PMode_UsesNewSession(t *testing.T) {
	rooms := makeSFURooms()
	rooms.config.SFUMode = false
	rooms.webrtcAPI = nil // must stay nil in P2P mode

	room := makeTestRoom("r", ConnectionLocal)
	rooms.Rooms["r"] = room
	rooms.connected[xid.New()] = "r" // host connected

	host, _ := makeTestUser(true)
	viewer, _ := makeTestUser(false)
	hostID := host.ID
	room.Users[hostID] = host
	room.Users[viewer.ID] = viewer
	rooms.connected[hostID] = "r"
	rooms.connected[viewer.ID] = "r"

	host.Streaming = false
	ev := &StartShare{}
	require.NoError(t, ev.Execute(rooms, ClientInfo{ID: hostID}))

	// P2P: one session created per viewer.
	assert.Len(t, room.Sessions, 1)
	// SFU fields must remain untouched.
	assert.Empty(t, room.SFUHosts)
}
