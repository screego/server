package mode

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	mode = Prod
	require.Equal(t, Prod, Get())
}

func TestSet(t *testing.T) {
	Set(Prod)
	require.Equal(t, Prod, mode)
}
