package ipdns

import "net"

type Provider interface {
	Get() (net.IP, net.IP, error)
}
