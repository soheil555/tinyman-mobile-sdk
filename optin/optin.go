package optin

import (
	"tinyman-mobile-sdk/utils"

	"github.com/algorand/go-algorand-sdk/future"
	"github.com/algorand/go-algorand-sdk/types"
)

func PrepareAppOptinTransactions(validatorAppId uint64, sender types.Address, suggestedParams types.SuggestedParams) (utils.TransactionGroup, error) {

	txn, err := future.MakeApplicationOptInTx(validatorAppId, nil, nil, nil, nil, suggestedParams, sender, nil, types.Digest{}, [32]byte{}, types.Address{})

	if err != nil {
		return utils.TransactionGroup{}, err
	}

	transactions := []types.Transaction{txn}

	txnGroup, err := utils.NewTransactionGroup(transactions)

	if err != nil {
		return utils.TransactionGroup{}, err
	}

	return txnGroup, nil

}

func PrepareAssetOptinTransactions(assetID uint64, sender types.Address, suggestedParams types.SuggestedParams) (utils.TransactionGroup, error) {

	//TODO: is it the same as AssetOptInTxn
	txn, err := future.MakeAssetTransferTxn(sender.String(), sender.String(), 0, nil, suggestedParams, "", assetID)
	if err != nil {
		return utils.TransactionGroup{}, err
	}

	transactions := []types.Transaction{txn}

	txnGroup, err := utils.NewTransactionGroup(transactions)

	if err != nil {
		return utils.TransactionGroup{}, err
	}

	return txnGroup, nil

}
