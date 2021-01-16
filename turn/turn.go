package turn

import (
	"net"

	"github.com/rs/xid"
)

type Account struct {
	Id         xid.ID
	Username   string
	Credential string
	IP         net.IP
}

type TurnServer interface {
	AcceptAccounts(...*Account) error
	RevokeAccounts(...xid.ID)
	Port() int
}
