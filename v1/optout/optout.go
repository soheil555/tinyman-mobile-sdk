package optout

import (
	"github.com/soheil555/tinyman-mobile-sdk/utils"
	"github.com/soheil555/tinyman-mobile-sdk/v1/client"

	"github.com/algorand/go-algorand-sdk/future"
	algoTypes "github.com/algorand/go-algorand-sdk/types"
)

func GetOptoutTransactions(client *client.TinymanClient, senderAddress string, validatorAppId int) (trxGroup *utils.TransactionGroup, err error) {

	sender, err := algoTypes.DecodeAddress(senderAddress)
	if err != nil {
		return
	}

	suggestedParams, err := client.SuggestedParams()

	if err != nil {
		return
	}

	txn, err := future.MakeApplicationClearStateTx(uint64(validatorAppId), nil, nil, nil, nil, suggestedParams, sender, nil, algoTypes.Digest{}, [32]byte{}, algoTypes.Address{})

	if err != nil {
		return
	}

	transactions := []algoTypes.Transaction{txn}

	trxGroup, err = utils.NewTransactionGroup(transactions)

	return
}
