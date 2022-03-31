package client

import (
	"reflect"
	"tinyman-mobile-sdk/constants"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/client/v2/common"
	"github.com/algorand/go-algorand-sdk/types"
)

type TinymanClient struct {
	Algod          *algod.Client
	ValidatorAppId uint64
	AssetsCache    map[uint64]interface{}
	UserAddress    types.Address
}

func NewTinymanClient(algodClient *algod.Client, validatorAppId uint64, userAddress types.Address) TinymanClient {

	return TinymanClient{
		algodClient,
		validatorAppId,
		map[uint64]interface{}{},
		userAddress,
	}
}

func NewTinymanTestnetClient(algodClient *algod.Client, userAddress types.Address) (TinymanClient, error) {

	//TODO: better way
	if reflect.DeepEqual(algodClient, algod.Client{}) {

		headers := []*common.Header{
			{Key: "User-Agent", Value: "algosdk"},
		}

		var err error

		algodClient, err = algod.MakeClientWithHeaders("https://api.testnet.algoexplorer.io", "", headers)

		if err != nil {
			return TinymanClient{}, err
		}

	}

	return NewTinymanClient(algodClient, constants.TESTNET_VALIDATOR_APP_ID, userAddress), nil

}

func NewTinymanMainnetClient(algodClient *algod.Client, userAddress types.Address) (TinymanClient, error) {

	//TODO: better way
	if reflect.DeepEqual(algodClient, algod.Client{}) {

		headers := []*common.Header{
			{Key: "User-Agent", Value: "algosdk"},
		}

		var err error
		algodClient, err = algod.MakeClientWithHeaders("https://api.algoexplorer.io", "", headers)

		if err != nil {
			return TinymanClient{}, err
		}

	}

	return NewTinymanClient(algodClient, constants.MAINNET_VALIDATOR_APP_ID, userAddress), nil

}

//TODO: implement later
func (s *TinymanClient) FetchPool(asset1 interface{}, asset2 interface{}, fetch bool) {

}
