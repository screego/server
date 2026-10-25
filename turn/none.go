package turn

import (
	"errors"
	"fmt"
	"net"

	"github.com/pion/turn/v5"
)

type RelayAddressGeneratorNone struct{}

func (r *RelayAddressGeneratorNone) Validate() error {
	return nil
}

func (r *RelayAddressGeneratorNone) AllocatePacketConn(conf turn.AllocateListenerConfig) (net.PacketConn, net.Addr, error) {
	conn, err := net.ListenPacket("udp", fmt.Sprintf(":%d", conf.RequestedPort))
	if err != nil {
		return nil, nil, err
	}

	return conn, conn.LocalAddr(), nil
}

func (r *RelayAddressGeneratorNone) AllocateListener(_ turn.AllocateListenerConfig) (net.Listener, net.Addr, error) {
	return nil, nil, errors.New("TCP relay not supported")
}

func (r *RelayAddressGeneratorNone) AllocateConn(_ turn.AllocateConnConfig) (net.Conn, error) {
	return nil, errors.New("TCP relay not supported")
}
