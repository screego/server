package turn

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/pion/turn/v2"
	"github.com/rs/zerolog/log"
	"github.com/screego/server/config"
	"github.com/screego/server/util"
)

type Server interface {
	Credentials(id string, addr net.IP) (string, string)
	Disallow(username string)
}

type InternalServer struct {
	lock       sync.RWMutex
	strictAuth bool
	lookup     map[string]Entry
}

type ExternalServer struct {
	secret []byte
	ttl    time.Duration
}

type Entry struct {
	addr     net.IP
	password []byte
}

const Realm = "screego"

type Generator struct {
	ipv4 net.IP
	ipv6 net.IP
	turn.RelayAddressGenerator
}

func (r *Generator) AllocatePacketConn(network string, requestedPort int) (net.PacketConn, net.Addr, error) {
	conn, addr, err := r.RelayAddressGenerator.AllocatePacketConn(network, requestedPort)
	if err != nil {
		return conn, addr, err
	}
	relayAddr := *addr.(*net.UDPAddr)
	if r.ipv6 == nil || (relayAddr.IP.To4() != nil && r.ipv4 != nil) {
		relayAddr.IP = r.ipv4
	} else {
		relayAddr.IP = r.ipv6
	}
	if err == nil {
		log.Debug().Str("addr", addr.String()).Str("relayaddr", relayAddr.String()).Msg("TURN allocated")
	}
	return conn, &relayAddr, err
}

func Start(conf config.Config) (Server, error) {
	if conf.TurnExternal {
		return newExternalServer(conf)
	} else {
		return newInternalServer(conf)
	}
}

func newExternalServer(conf config.Config) (Server, error) {
	return &ExternalServer{
		secret: []byte(conf.TurnExternalSecret),
		ttl:    24 * time.Hour,
	}, nil
}

func newInternalServer(conf config.Config) (Server, error) {
	udpListener, err := net.ListenPacket("udp", conf.TurnAddress)
	if err != nil {
		return nil, fmt.Errorf("udp: could not listen on %s: %s", conf.TurnAddress, err)
	}
	tcpListener, err := net.Listen("tcp", conf.TurnAddress)
	if err != nil {
		return nil, fmt.Errorf("tcp: could not listen on %s: %s", conf.TurnAddress, err)
	}

	svr := &InternalServer{
		lookup:     map[string]Entry{},
		strictAuth: conf.TurnStrictAuth,
	}

	gen := &Generator{
		ipv4:                  conf.TurnIPV4,
		ipv6:                  conf.TurnIPV6,
		RelayAddressGenerator: generator(conf),
	}

	_, err = turn.NewServer(turn.ServerConfig{
		Realm:       Realm,
		AuthHandler: svr.authenticate,
		ListenerConfigs: []turn.ListenerConfig{
			{Listener: tcpListener, RelayAddressGenerator: gen},
		},
		PacketConnConfigs: []turn.PacketConnConfig{
			{PacketConn: udpListener, RelayAddressGenerator: gen},
		},
	})
	if err != nil {
		return nil, err
	}

	log.Info().Str("addr", conf.TurnAddress).Msg("Start TURN/STUN")
	return svr, nil
}

func generator(conf config.Config) turn.RelayAddressGenerator {
	min, max, useRange := conf.PortRange()
	if useRange {
		log.Debug().Uint16("min", min).Uint16("max", max).Msg("Using Port Range")
		return &RelayAddressGeneratorPortRange{MinPort: min, MaxPort: max}
	}
	return &RelayAddressGeneratorNone{}
}

func (a *InternalServer) allow(username, password string, addr net.IP) {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.lookup[username] = Entry{
		addr:     addr,
		password: turn.GenerateAuthKey(username, Realm, password),
	}
}

func (a *InternalServer) Disallow(username string) {
	a.lock.Lock()
	defer a.lock.Unlock()

	delete(a.lookup, username)
}

func (a *ExternalServer) Disallow(username string) {
	// not supported, will expire on TTL
}

func (a *InternalServer) authenticate(username, realm string, addr net.Addr) ([]byte, bool) {
	a.lock.RLock()
	defer a.lock.RUnlock()

	var connectedIp net.IP
	switch addr := addr.(type) {
	case *net.UDPAddr:
		connectedIp = addr.IP
	case *net.TCPAddr:
		connectedIp = addr.IP
	default:
		log.Error().Interface("type", fmt.Sprintf("%T", addr)).Msg("unknown addr type")
		return nil, false
	}
	entry, ok := a.lookup[username]

	if !ok {
		log.Debug().Interface("addr", addr).Str("username", username).Msg("TURN username not found")
		return nil, false
	}

	authIP := entry.addr

	if a.strictAuth && !connectedIp.Equal(authIP) {
		log.Debug().Interface("allowedIp", addr.String()).Interface("connectingIp", entry.addr.String()).Msg("TURN strict auth check failed")
		return nil, false
	}

	log.Debug().Interface("addr", addr.String()).Str("realm", realm).Msg("TURN authenticated")
	return entry.password, true
}

func (a *InternalServer) Credentials(id string, addr net.IP) (string, string) {
	password := util.RandString(20)
	a.allow(id, password, addr)
	return id, password
}

func (a *ExternalServer) Credentials(id string, addr net.IP) (string, string) {
	username := fmt.Sprintf("%d:%s", time.Now().Add(a.ttl).Unix(), id)
	mac := hmac.New(sha1.New, a.secret)
	_, _ = mac.Write([]byte(username))
	password := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return username, password
}
