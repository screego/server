package turn

import "net"

type TurnServer interface {
	Allow(username, password string, addr net.IP)
	Disallow(username string)
	Port() int
}
