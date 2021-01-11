package turn

import (
	"crypto/hmac"
	"crypto/sha1"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPassword(t *testing.T) {
	secret := []byte("s3cr37")
	user, password, err := buildPassword("bob", 15*time.Minute, secret)
	assert.NoError(t, err)
	split := strings.Split(user, ":")
	assert.Len(t, split, 2)
	assert.Equal(t, "bob", split[1])
	u, err := strconv.Atoi(split[0])
	assert.NoError(t, err)
	assert.True(t, int64(u) > time.Now().Unix())
	mac := hmac.New(sha1.New, secret)
	mac.Write([]byte(user))
	assert.Equal(t, string(mac.Sum(nil)), password)
}
