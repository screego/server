package turn

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rs/xid"
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
	_, err := mac.Write([]byte(account.Username))
	if err != nil {
		return err
	}
	account.Credential = base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return nil
}

func (t *TurnREST) RevokeAccounts(...xid.ID) {
	// do nothing, wait for peremption
}

func (t *TurnREST) Port() int {
	return t.port
}

func newTurnREST(conf config.Config) (TurnServer, error) {
	if conf.TurnCoturnSignedSecret == "" {
		return nil, errors.New("TurnCoturnSignedSecret can't be empty")
	}
	slugs := strings.Split(conf.TurnAddress, ":")
	if len(slugs) == 0 {
		return nil, fmt.Errorf("Can't read CoturnAddress : %s", conf.TurnAddress)
	}
	port, err := strconv.Atoi(slugs[len(slugs)-1])
	if err != nil {
		return nil, fmt.Errorf("Can't parse TurnAddress port : %s %s", slugs[len(slugs)-1], err)
	}
	return &TurnREST{
		ttl:    24 * time.Hour, // For now, it's hardcoded
		secret: []byte(conf.TurnCoturnSignedSecret),
		port:   port,
	}, nil
}
