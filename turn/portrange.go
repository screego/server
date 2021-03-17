package turn

import (
	"errors"
	"fmt"
	"net"

	"github.com/pion/randutil"
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

func (r *RelayAddressGeneratorPortRange) AllocatePacketConn(network string, requestedPort int) (net.PacketConn, net.Addr, error) {
	if requestedPort != 0 {
		conn, err := net.ListenPacket("udp", fmt.Sprintf(":%d", requestedPort))
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

func (r *RelayAddressGeneratorPortRange) AllocateConn(network string, requestedPort int) (net.Conn, net.Addr, error) {
	return nil, nil, errors.New("todo")
}
