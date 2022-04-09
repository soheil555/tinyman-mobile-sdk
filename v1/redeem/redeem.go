package redeem

import (
	"tinyman-mobile-sdk/utils"
	"tinyman-mobile-sdk/v1/contracts"

	"github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/future"
	"github.com/algorand/go-algorand-sdk/types"
)

func PrepareRedeemTransactions(validatorAppId uint64, asset1ID uint64, asset2ID uint64, liquidityAssetID uint64, assetID uint64, assetAmount uint64, sender types.Address, suggestedParams types.SuggestedParams) (utils.TransactionGroup, error) {

	poolLogicsig, err := contracts.GetPoolLogicsig(validatorAppId, asset1ID, asset2ID)

	if err != nil {
		return utils.TransactionGroup{}, err
	}

	poolAddress := crypto.AddressFromProgram(poolLogicsig.Logic)

	paymentTxn, err := future.MakePaymentTxn(sender.String(), poolAddress.String(), 2000, []byte("fee"), "", suggestedParams)

	if err != nil {
		return utils.TransactionGroup{}, err
	}

	var foreignAssets []uint64

	if asset2ID == 0 {
		foreignAssets = []uint64{asset1ID, liquidityAssetID}
	} else {
		foreignAssets = []uint64{asset1ID, asset2ID, liquidityAssetID}
	}

	applicationNoOptTxn, err := future.MakeApplicationNoOpTx(validatorAppId, [][]byte{[]byte("redeem")}, []string{sender.String()}, nil, foreignAssets, suggestedParams, poolAddress, nil, types.Digest{}, [32]byte{}, types.Address{})

	if err != nil {
		return utils.TransactionGroup{}, err
	}

	var assetTransferTxn types.Transaction

	if assetID != 0 {

		assetTransferTxn, err = future.MakeAssetTransferTxn(poolAddress.String(), sender.String(), assetAmount, nil, suggestedParams, "", assetID)

	} else {

		assetTransferTxn, err = future.MakePaymentTxn(poolAddress.String(), sender.String(), assetAmount, nil, "", suggestedParams)

	}

	if err != nil {
		return utils.TransactionGroup{}, err
	}

	txns := []types.Transaction{paymentTxn, applicationNoOptTxn, assetTransferTxn}

	txnGroup, err := utils.NewTransactionGroup(txns)

	if err != nil {
		return utils.TransactionGroup{}, err
	}

	err = txnGroup.SignWithLogicsig(poolLogicsig)
	if err != nil {
		return utils.TransactionGroup{}, err
	}

	return txnGroup, nil

}
