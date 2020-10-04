package mode

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestGet(t *testing.T) {
	mode = Prod
	assert.Equal(t, Prod, Get())
}

func TestSet(t *testing.T) {
	Set(Prod)
	assert.Equal(t, Prod, mode)
}
