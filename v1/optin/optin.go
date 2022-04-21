package optin

import (
	"tinyman-mobile-sdk/utils"

	"github.com/algorand/go-algorand-sdk/future"
	algoTypes "github.com/algorand/go-algorand-sdk/types"
)

func PrepareAppOptinTransactions(validatorAppId uint64, sender algoTypes.Address, suggestedParams algoTypes.SuggestedParams) (txnGroup utils.TransactionGroup, err error) {

	txn, err := future.MakeApplicationOptInTx(validatorAppId, nil, nil, nil, nil, suggestedParams, sender, nil, algoTypes.Digest{}, [32]byte{}, algoTypes.Address{})

	if err != nil {
		return
	}

	transactions := []algoTypes.Transaction{txn}

	txnGroup, err = utils.NewTransactionGroup(transactions)

	return

}

func PrepareAssetOptinTransactions(assetID uint64, sender algoTypes.Address, suggestedParams algoTypes.SuggestedParams) (txnGroup utils.TransactionGroup, err error) {

	txn, err := future.MakeAssetTransferTxn(sender.String(), sender.String(), 0, nil, suggestedParams, "", assetID)
	if err != nil {
		return
	}

	transactions := []algoTypes.Transaction{txn}

	txnGroup, err = utils.NewTransactionGroup(transactions)

	return

}
