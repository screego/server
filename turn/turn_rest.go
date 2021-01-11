package turn

import (
	"crypto/hmac"
	"crypto/sha1"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/screego/server/config"
)

type TurnREST struct {
}

func (t *TurnREST) Allow(username, password string, addr net.IP) {
	//FIXME
}

func buildPassword(my_user string, ttl time.Duration, secret []byte) (user, password string, err error) {
	//https://stackoverflow.com/questions/35766382/coturn-how-to-use-turn-rest-api#54725092
	if ttl <= 0 {
		return "", "", errors.New("Use a TTL > 0")
	}
	user = fmt.Sprintf("%d:%s", time.Now().Add(ttl).Unix(), my_user)
	mac := hmac.New(sha1.New, secret)
	mac.Write([]byte(user))
	password = string(mac.Sum(nil))
	return user, password, nil
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
