package optout

import (
	"context"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/future"
	algoTypes "github.com/algorand/go-algorand-sdk/types"
)

func GetOptoutTransactions(client *algod.Client, sender []byte, validatorAppId int) (transactions []*algoTypes.Transaction, err error) {

	var senderAddress algoTypes.Address
	copy(senderAddress[:], sender)

	suggestedParams, err := client.SuggestedParams().Do(context.Background())

	if err != nil {
		return
	}

	txn, err := future.MakeApplicationClearStateTx(uint64(validatorAppId), nil, nil, nil, nil, suggestedParams, senderAddress, nil, algoTypes.Digest{}, [32]byte{}, algoTypes.Address{})

	if err != nil {
		return
	}

	transactions = []*algoTypes.Transaction{&txn}

	return
}
