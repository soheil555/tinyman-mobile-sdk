package contracts

import (
	"testing"

	"github.com/soheil555/tinyman-mobile-sdk/types"

	"github.com/algorand/go-algorand-sdk/crypto"
	"github.com/stretchr/testify/assert"
)

func TestReadContractsFile(t *testing.T) {

	contracts, err := readContractsFile()
	assert.Nil(t, err)

	assert.NotEqual(t, types.ASC{}, contracts)

}

func TestGetPoolLogicsig(t *testing.T) {

	var validatorAppID int = 1
	var asset1ID int = 1
	var asset2ID int = 2

	expectedAddress := "7ZRYUGMMMGCBBQYMKEHIU7YMZ7WW6H4ADOIBAH3MCELK3KGAUC7MVJ5OAY"

	lsig, err := GetPoolLogicsig(validatorAppID, asset1ID, asset2ID)
	assert.Nil(t, err)

	actualAddress := crypto.AddressFromProgram(lsig.Logic).String()

	assert.Equal(t, expectedAddress, actualAddress)

}
