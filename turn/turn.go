package turn

import (
	"net"

	"github.com/rs/xid"
)

type TurnAccount struct {
	Id         xid.ID
	Username   string
	Credential string
	IP         net.IP
}

type TurnServer interface {
	AcceptAccounts(client, host *TurnAccount) error
	RevokeAccounts(client, host *TurnAccount)
	Port() int
}
