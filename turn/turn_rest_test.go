package turn

import (
	"crypto/hmac"
	"crypto/sha1"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
)

func TestPassword(t *testing.T) {
	secret := []byte("s3cr37")
	account := &TurnAccount{
		Id: xid.New(),
	}
	err := buildPassword(account, 15*time.Minute, secret)
	assert.NoError(t, err)
	split := strings.Split(account.Username, ":")
	assert.Len(t, split, 2)
	assert.Equal(t, account.Id.String(), split[1])
	u, err := strconv.Atoi(split[0])
	assert.NoError(t, err)
	assert.True(t, int64(u) > time.Now().Unix())
	mac := hmac.New(sha1.New, secret)
	_, err = mac.Write([]byte(account.Username))
	assert.NoError(t, err)
	assert.Equal(t, string(mac.Sum(nil)), account.Credential)
}
