package redeem

import (
	"tinyman-mobile-sdk/utils"
	"tinyman-mobile-sdk/v1/contracts"

	"github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/future"
	algoTypes "github.com/algorand/go-algorand-sdk/types"
)

func PrepareRedeemTransactions(validatorAppId uint64, asset1ID uint64, asset2ID uint64, liquidityAssetID uint64, assetID uint64, assetAmount uint64, sender algoTypes.Address, suggestedParams algoTypes.SuggestedParams) (txnGroup utils.TransactionGroup, err error) {

	poolLogicsig, err := contracts.GetPoolLogicsig(validatorAppId, asset1ID, asset2ID)

	if err != nil {
		return
	}

	poolAddress := crypto.AddressFromProgram(poolLogicsig.Logic)

	paymentTxn, err := future.MakePaymentTxn(sender.String(), poolAddress.String(), 2000, []byte("fee"), "", suggestedParams)

	if err != nil {
		return
	}

	var foreignAssets []uint64

	if asset2ID == 0 {
		foreignAssets = []uint64{asset1ID, liquidityAssetID}
	} else {
		foreignAssets = []uint64{asset1ID, asset2ID, liquidityAssetID}
	}

	applicationNoOptTxn, err := future.MakeApplicationNoOpTx(validatorAppId, [][]byte{[]byte("redeem")}, []string{sender.String()}, nil, foreignAssets, suggestedParams, poolAddress, nil, algoTypes.Digest{}, [32]byte{}, algoTypes.Address{})

	if err != nil {
		return
	}

	var assetTransferTxn algoTypes.Transaction

	if assetID != 0 {

		assetTransferTxn, err = future.MakeAssetTransferTxn(poolAddress.String(), sender.String(), assetAmount, nil, suggestedParams, "", assetID)

	} else {

		assetTransferTxn, err = future.MakePaymentTxn(poolAddress.String(), sender.String(), assetAmount, nil, "", suggestedParams)

	}

	if err != nil {
		return
	}

	txns := []algoTypes.Transaction{paymentTxn, applicationNoOptTxn, assetTransferTxn}

	txnGroup, err = utils.MakeTransactionGroup(txns)

	if err != nil {
		return
	}

	err = txnGroup.SignWithLogicsig(poolLogicsig)
	if err != nil {
		return
	}

	return

}
