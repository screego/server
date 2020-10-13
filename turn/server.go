package turn

import (
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/pion/turn/v2"
	"github.com/rs/zerolog/log"
	"github.com/screego/server/config"
)

type Server struct {
	TurnAddress   string
	StunAddress   string
	lock          sync.RWMutex
	strictIPCheck bool
	lookup        map[string]Entry
}

type Entry struct {
	addr     net.IP
	password []byte
}

const Realm = "screego"

type LoggedGenerator struct {
	turn.RelayAddressGenerator
}

func (r *LoggedGenerator) AllocatePacketConn(network string, requestedPort int) (net.PacketConn, net.Addr, error) {
	conn, addr, err := r.RelayAddressGenerator.AllocatePacketConn(network, requestedPort)
	if err == nil {
		log.Debug().Str("addr", addr.String()).Str("network", network).Msg("TURN allocated")
	}
	return conn, addr, err
}

func Start(conf config.Config) (*Server, error) {
	udpListener, err := net.ListenPacket("udp4", conf.TurnAddress)
	if err != nil {
		return nil, fmt.Errorf("udp: could not listen on %s: %s", conf.TurnAddress, err)
	}
	tcpListener, err := net.Listen("tcp4", conf.TurnAddress)
	if err != nil {
		return nil, fmt.Errorf("tcp: could not listen on %s: %s", conf.TurnAddress, err)
	}

	split := strings.SplitN(conf.TurnAddress, ":", 2)
	svr := &Server{
		TurnAddress:   fmt.Sprintf("turn:%s:%s", conf.ExternalIP, split[1]),
		StunAddress:   fmt.Sprintf("stun:%s:%s", conf.ExternalIP, split[1]),
		lookup:        map[string]Entry{},
		strictIPCheck: conf.TurnStrictAuth,
	}

	loggedGenerator := &LoggedGenerator{RelayAddressGenerator: generator(conf)}

	_, err = turn.NewServer(turn.ServerConfig{
		Realm:       Realm,
		AuthHandler: svr.authenticate,
		ListenerConfigs: []turn.ListenerConfig{
			{Listener: tcpListener, RelayAddressGenerator: loggedGenerator},
		},
		PacketConnConfigs: []turn.PacketConnConfig{
			{PacketConn: udpListener, RelayAddressGenerator: loggedGenerator},
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
		return &turn.RelayAddressGeneratorPortRange{
			RelayAddress: net.ParseIP(conf.ExternalIP),
			Address:      "0.0.0.0",
			MinPort:      min,
			MaxPort:      max,
		}
	}
	return &turn.RelayAddressGeneratorStatic{
		RelayAddress: net.ParseIP(conf.ExternalIP),
		Address:      "0.0.0.0",
	}
}

func (a *Server) Allow(username, password string, addr net.IP) {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.lookup[username] = Entry{
		addr:     addr,
		password: turn.GenerateAuthKey(username, Realm, password),
	}
}

func (a *Server) Disallow(username string) {
	a.lock.Lock()
	defer a.lock.Unlock()

	delete(a.lookup, username)
}

func (a *Server) authenticate(username, realm string, addr net.Addr) ([]byte, bool) {
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

	if !connectedIp.Equal(authIP) {
		if a.strictIPCheck {
			log.Debug().Interface("allowedIp", addr.String()).Interface("connectingIp", entry.addr.String()).Msg("TURN strict ip check failed")
			return nil, false
		}

		conIPIsV4 := connectedIp.To4() != nil
		authIPIsV4 := authIP.To4() != nil

		if authIPIsV4 == conIPIsV4 {
			log.Debug().Interface("allowedIp", addr.String()).Interface("connectingIp", entry.addr.String()).Msg("TURN ip check failed")
			return nil, false
		}
	}
	log.Debug().Interface("addr", addr.String()).Str("realm", realm).Msg("TURN authenticated")
	return entry.password, true
}
