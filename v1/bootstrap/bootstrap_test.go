package bootstrap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHex2Int(t *testing.T) {

	var expected uint64 = 18446744073709551615
	actual := hex2Int("0xffffffffffffffff")

	assert.Equal(t, expected, actual)

}
