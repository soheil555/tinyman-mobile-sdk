package optout

import (
	"context"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/future"
	algoTypes "github.com/algorand/go-algorand-sdk/types"
)

func GetOptoutTransactions(client algod.Client, sender algoTypes.Address, validatorAppId uint64) ([]algoTypes.Transaction, error) {

	suggestedParams, err := client.SuggestedParams().Do(context.Background())

	if err != nil {
		return nil, err
	}

	txn, err := future.MakeApplicationClearStateTx(validatorAppId, nil, nil, nil, nil, suggestedParams, sender, nil, algoTypes.Digest{}, [32]byte{}, algoTypes.Address{})

	if err != nil {
		return nil, err
	}

	return []algoTypes.Transaction{txn}, nil
}
