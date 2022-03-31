package swap

import (
	"crypto/ed25519"
	"fmt"
	"tinyman-mobile-sdk/contracts"
	"tinyman-mobile-sdk/utils"

	"github.com/algorand/go-algorand-sdk/future"
	"github.com/algorand/go-algorand-sdk/logic"
	"github.com/algorand/go-algorand-sdk/types"
)

func PrepareSwapTransactions(validatorAppId uint64, asset1ID uint64, asset2ID uint64, liquidityAssetID uint64, assetInID uint64, assetInAmount uint64, assetOutAmount uint64, swapType string, sender types.Address, suggestedParams types.SuggestedParams) (utils.TransactionGroup, error) {

	poolLogicsig, err := contracts.GetPoolLogicsig(validatorAppId, asset1ID, asset2ID)

	if err != nil {
		return utils.TransactionGroup{}, err
	}

	//TODO: what is pool address
	_, byteArrays, err := logic.ReadProgram(poolLogicsig.Logic, nil)

	if err != nil {
		return utils.TransactionGroup{}, err
	}

	//TODO: where is address in byteArray?
	var poolAddress types.Address

	n := copy(poolAddress[:], byteArrays[1])

	if n != ed25519.PublicKeySize {
		return utils.TransactionGroup{}, fmt.Errorf("address generated from receiver bytes is the wrong size")
	}

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

	applicationNoOptTxn, err := future.MakeApplicationNoOpTx(validatorAppId, [][]byte{[]byte("swap"), []byte(swapTypes[swapType])}, []string{sender.String()}, nil, foreignAssets, suggestedParams, poolAddress, nil, types.Digest{}, [32]byte{}, types.Address{})

	if err != nil {
		return utils.TransactionGroup{}, err
	}

	var assetTransferInTxn types.Transaction

	if assetInID != 0 {
		assetTransferInTxn, err = future.MakeAssetTransferTxn(sender.String(), poolAddress.String(), assetInAmount, nil, suggestedParams, "", assetInID)
	} else {
		assetTransferInTxn, err = future.MakePaymentTxn(sender.String(), poolAddress.String(), assetInAmount, nil, "", suggestedParams)
	}

	if err != nil {
		return utils.TransactionGroup{}, err
	}

	var assetTransferOutTxn types.Transaction

	if assetOutID != 0 {
		assetTransferOutTxn, err = future.MakeAssetTransferTxn(poolAddress.String(), sender.String(), assetOutAmount, nil, suggestedParams, "", assetOutID)
	} else {
		assetTransferOutTxn, err = future.MakePaymentTxn(poolAddress.String(), sender.String(), assetOutAmount, nil, "", suggestedParams)
	}

	if err != nil {
		return utils.TransactionGroup{}, err
	}

	txns := []types.Transaction{paymentTxn, applicationNoOptTxn, assetTransferInTxn, assetTransferOutTxn}

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
