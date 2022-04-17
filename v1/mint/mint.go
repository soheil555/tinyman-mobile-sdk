package mint

import (
	"tinyman-mobile-sdk/utils"
	"tinyman-mobile-sdk/v1/contracts"

	"github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/future"
	algoTypes "github.com/algorand/go-algorand-sdk/types"
)

func PrepareMintTransactions(validatorAppId uint64, asset1ID uint64, asset2ID uint64, liquidityAssetID uint64, asset1Amount uint64, asset2Amount uint64, liquidityAssetAmount uint64, sender algoTypes.Address, suggestedParams algoTypes.SuggestedParams) (txnGroup utils.TransactionGroup, err error) {

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

	applicationNoOpTxn, err := future.MakeApplicationNoOpTx(validatorAppId, [][]byte{[]byte("mint")}, []string{sender.String()}, nil, foreignAssets, suggestedParams, poolAddress, nil, algoTypes.Digest{}, [32]byte{}, algoTypes.Address{})

	if err != nil {
		return
	}

	assetTransferTxn1, err := future.MakeAssetTransferTxn(sender.String(), poolAddress.String(), asset1Amount, nil, suggestedParams, "", asset1ID)

	if err != nil {
		return
	}

	var assetTransferTxn2 algoTypes.Transaction

	if asset2ID != 0 {
		assetTransferTxn2, err = future.MakeAssetTransferTxn(sender.String(), poolAddress.String(), asset2Amount, nil, suggestedParams, "", asset2ID)
	} else {
		assetTransferTxn2, err = future.MakePaymentTxn(sender.String(), poolAddress.String(), asset2Amount, nil, "", suggestedParams)
	}

	if err != nil {
		return
	}

	assetTransferTxn3, err := future.MakeAssetTransferTxn(poolAddress.String(), sender.String(), liquidityAssetAmount, nil, suggestedParams, "", liquidityAssetID)

	if err != nil {
		return
	}

	txns := []algoTypes.Transaction{paymentTxn, applicationNoOpTxn, assetTransferTxn1, assetTransferTxn2, assetTransferTxn3}

	txnGroup, err = utils.MakeTransactionGroup(txns)

	if err != nil {
		return
	}

	err = txnGroup.SignWithLogicsig(poolLogicsig)

	return

}
