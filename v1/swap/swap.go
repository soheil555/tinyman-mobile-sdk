package swap

import (
	"tinyman-mobile-sdk/utils"
	"tinyman-mobile-sdk/v1/contracts"

	"github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/future"
	algoTypes "github.com/algorand/go-algorand-sdk/types"
)

func PrepareSwapTransactions(validatorAppId uint64, asset1ID uint64, asset2ID uint64, liquidityAssetID uint64, assetInID uint64, assetInAmount uint64, assetOutAmount uint64, swapType string, sender algoTypes.Address, suggestedParams algoTypes.SuggestedParams) (utils.TransactionGroup, error) {

	poolLogicsig, err := contracts.GetPoolLogicsig(validatorAppId, asset1ID, asset2ID)

	if err != nil {
		return utils.TransactionGroup{}, err
	}

	poolAddress := crypto.AddressFromProgram(poolLogicsig.Logic)

	swapTypes := map[string]string{
		"fixed-input":  "fi",
		"fixed-output": "fo",
	}

	var assetOutID uint64

	if assetInID == asset1ID {
		assetOutID = asset2ID
	} else {
		assetOutID = asset1ID
	}

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

	applicationNoOptTxn, err := future.MakeApplicationNoOpTx(validatorAppId, [][]byte{[]byte("swap"), []byte(swapTypes[swapType])}, []string{sender.String()}, nil, foreignAssets, suggestedParams, poolAddress, nil, algoTypes.Digest{}, [32]byte{}, algoTypes.Address{})

	if err != nil {
		return utils.TransactionGroup{}, err
	}

	var assetTransferInTxn algoTypes.Transaction

	if assetInID != 0 {
		assetTransferInTxn, err = future.MakeAssetTransferTxn(sender.String(), poolAddress.String(), assetInAmount, nil, suggestedParams, "", assetInID)
	} else {
		assetTransferInTxn, err = future.MakePaymentTxn(sender.String(), poolAddress.String(), assetInAmount, nil, "", suggestedParams)
	}

	if err != nil {
		return utils.TransactionGroup{}, err
	}

	var assetTransferOutTxn algoTypes.Transaction

	if assetOutID != 0 {
		assetTransferOutTxn, err = future.MakeAssetTransferTxn(poolAddress.String(), sender.String(), assetOutAmount, nil, suggestedParams, "", assetOutID)
	} else {
		assetTransferOutTxn, err = future.MakePaymentTxn(poolAddress.String(), sender.String(), assetOutAmount, nil, "", suggestedParams)
	}

	if err != nil {
		return utils.TransactionGroup{}, err
	}

	txns := []algoTypes.Transaction{paymentTxn, applicationNoOptTxn, assetTransferInTxn, assetTransferOutTxn}

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
