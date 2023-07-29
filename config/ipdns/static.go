package ipdns

import "net"

type Static struct {
	V4 net.IP
	V6 net.IP
}

func (s *Static) Get() (net.IP, net.IP, error) {
	return s.V4, s.V6, nil
}
