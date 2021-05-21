package turn

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net"
	"strings"
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
	IPV4() net.IP
	IPV6() net.IP
	Port() string
}

type InternalServer struct {
	ipV4       net.IP
	ipV6       net.IP
	port       string
	lock       sync.RWMutex
	strictAuth bool
	lookup     map[string]Entry
}

type ExternalServer struct {
	ipV4   net.IP
	ipV6   net.IP
	port   string
	secret []byte
	ttl    time.Duration
}

type Entry struct {
	addr     net.IP
	password []byte
}

const Realm = "screego"

type Generator struct {
	ipV4 net.IP
	ipV6 net.IP
	turn.RelayAddressGenerator
}

func (r *Generator) AllocatePacketConn(network string, requestedPort int) (net.PacketConn, net.Addr, error) {
	conn, addr, err := r.RelayAddressGenerator.AllocatePacketConn(network, requestedPort)
	if err != nil {
		return conn, addr, err
	}
	relayAddr := *addr.(*net.UDPAddr)
	if r.ipV6 == nil || (relayAddr.IP.To4() != nil && r.ipV4 != nil) {
		relayAddr.IP = r.ipV4
	} else {
		relayAddr.IP = r.ipV6
	}
	if err == nil {
		log.Debug().Str("addr", addr.String()).Str("relayaddr", relayAddr.String()).Msg("TURN allocated")
	}
	return conn, &relayAddr, err
}

func Start(conf config.Config) (Server, error) {
	if conf.TurnExternalIPV4 != nil || conf.TurnExternalIPV6 != nil {
		return newExternalServer(conf)
	} else {
		return newInternalServer(conf)
	}
}

func newExternalServer(conf config.Config) (Server, error) {
	return &ExternalServer{
		ipV4:   conf.TurnExternalIPV4,
		ipV6:   conf.TurnExternalIPV6,
		port:   conf.TurnExternalPort,
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

	split := strings.Split(conf.TurnAddress, ":")
	svr := &InternalServer{
		ipV4:       conf.ExternalIPV4,
		ipV6:       conf.ExternalIPV6,
		port:       split[len(split)-1],
		lookup:     map[string]Entry{},
		strictAuth: conf.TurnStrictAuth,
	}

	gen := &Generator{
		ipV4:                  conf.ExternalIPV4,
		ipV6:                  conf.ExternalIPV6,
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

func (a *InternalServer) Allow(username, password string, addr net.IP) {
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
	a.Allow(id, password, addr)
	return id, password
}

func (a *ExternalServer) Credentials(id string, addr net.IP) (string, string) {
	username := fmt.Sprintf("%d:%s", time.Now().Add(a.ttl).Unix(), id)
	mac := hmac.New(sha1.New, a.secret)
	_, _ = mac.Write([]byte(username))
	password := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return username, password
}

func (a *InternalServer) IPV4() net.IP {
	return a.ipV4
}

func (a *ExternalServer) IPV4() net.IP {
	return a.ipV4
}

func (a *InternalServer) IPV6() net.IP {
	return a.ipV6
}

func (a *ExternalServer) IPV6() net.IP {
	return a.ipV6
}

func (a *InternalServer) Port() string {
	return a.port
}

func (a *ExternalServer) Port() string {
	return a.port
}
