package turn

import (
	"errors"
	"fmt"
	"net"

	"github.com/pion/randutil"
	"github.com/pion/turn/v5"
)

type RelayAddressGeneratorPortRange struct {
	MinPort uint16
	MaxPort uint16
	Rand    randutil.MathRandomGenerator
}

func (r *RelayAddressGeneratorPortRange) Validate() error {
	if r.Rand == nil {
		r.Rand = randutil.NewMathRandomGenerator()
	}

	return nil
}

func (r *RelayAddressGeneratorPortRange) AllocatePacketConn(conf turn.AllocateListenerConfig) (net.PacketConn, net.Addr, error) {
	if conf.RequestedPort != 0 {
		conn, err := net.ListenPacket("udp", fmt.Sprintf(":%d", conf.RequestedPort))
		if err != nil {
			return nil, nil, err
		}
		relayAddr := conn.LocalAddr().(*net.UDPAddr)
		return conn, relayAddr, nil
	}

	for try := 0; try < 10; try++ {
		port := r.MinPort + uint16(r.Rand.Intn(int((r.MaxPort+1)-r.MinPort)))
		conn, err := net.ListenPacket("udp", fmt.Sprintf(":%d", port))
		if err != nil {
			continue
		}

		relayAddr := conn.LocalAddr().(*net.UDPAddr)
		return conn, relayAddr, nil
	}

	return nil, nil, errors.New("could not find free port: max retries exceeded")
}

func (r *RelayAddressGeneratorPortRange) AllocateListener(_ turn.AllocateListenerConfig) (net.Listener, net.Addr, error) {
	return nil, nil, errors.New("TCP relay not supported")
}

func (r *RelayAddressGeneratorPortRange) AllocateConn(_ turn.AllocateConnConfig) (net.Conn, error) {
	return nil, errors.New("TCP relay not supported")
}
