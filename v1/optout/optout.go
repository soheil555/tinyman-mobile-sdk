package optout

import (
	"context"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/future"
	"github.com/algorand/go-algorand-sdk/types"
)

func GetOptoutTransactions(client algod.Client, sender types.Address, validatorAppId uint64) ([]types.Transaction, error) {

	suggestedParams, err := client.SuggestedParams().Do(context.Background())

	if err != nil {
		return nil, err
	}

	txn, err := future.MakeApplicationClearStateTx(validatorAppId, nil, nil, nil, nil, suggestedParams, sender, nil, types.Digest{}, [32]byte{}, types.Address{})

	if err != nil {
		return nil, err
	}

	return []types.Transaction{txn}, nil
}
