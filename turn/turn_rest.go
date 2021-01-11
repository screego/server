package turn

import (
	"net"

	"github.com/screego/server/config"
)

type TurnREST struct {
}

func (t *TurnREST) Allow(username, password string, addr net.IP) {
	//FIXME

}

func (t *TurnREST) Disallow(username string) {
	//FIXME

}

func (t *TurnREST) Port() int {
	//FIXME
	return 0
}

func newTurnREST(conf config.Config) (TurnServer, error) {
	//FIXME
	return &TurnREST{}, nil
}
