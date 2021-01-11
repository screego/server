package turn

import (
	"crypto/hmac"
	"crypto/sha1"
	"errors"
	"fmt"
	"time"

	"github.com/screego/server/config"
)

type TurnREST struct {
	ttl    time.Duration
	secret []byte
	port   int
}

func (t *TurnREST) AcceptAccounts(client, host *TurnAccount) error {
	err := buildPassword(client, t.ttl, t.secret)
	if err != nil {
		return err
	}
	return buildPassword(host, t.ttl, t.secret)
}

func buildPassword(account *TurnAccount, ttl time.Duration, secret []byte) error {
	//https://stackoverflow.com/questions/35766382/coturn-how-to-use-turn-rest-api#54725092
	if ttl <= 0 {
		return errors.New("Use a TTL > 0")
	}
	account.Username = fmt.Sprintf("%d:%s", time.Now().Add(ttl).Unix(), account.Id.String())
	mac := hmac.New(sha1.New, secret)
	mac.Write([]byte(account.Username))
	account.Credential = string(mac.Sum(nil))
	return nil
}

func (t *TurnREST) RevokeAccounts(client, host *TurnAccount) {
	// do nothing, wait for peremption
}

func (t *TurnREST) Port() int {
	return t.port
}

func newTurnREST(conf config.Config) (TurnServer, error) {
	//FIXME
	return &TurnREST{
		ttl:    12 * time.Hour,
		secret: []byte("s3cr37"),
		port:   3478,
	}, nil
}
