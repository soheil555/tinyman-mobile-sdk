package contracts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPoolLogicsig(t *testing.T) {

	_, err := GetPoolLogicsig(21580889, 0, 2)
	assert.Nil(t, err)

	// fmt.Println(logicsig.Logic)

}
