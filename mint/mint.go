package mint

import (
	"crypto/ed25519"
	"fmt"
	"tinyman-mobile-sdk/contracts"
	"tinyman-mobile-sdk/utils"

	"github.com/algorand/go-algorand-sdk/future"
	"github.com/algorand/go-algorand-sdk/logic"
	"github.com/algorand/go-algorand-sdk/types"
)

func PrepareMintTransactions(validatorAppId uint64, asset1ID uint64, asset2ID uint64, liquidityAssetID uint64, asset1Amount uint64, asset2Amount uint64, liquidityAssetAmount uint64, sender types.Address, suggestedParams types.SuggestedParams) (utils.TransactionGroup, error) {

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

	applicationNoOptTxn, err := future.MakeApplicationNoOpTx(validatorAppId, [][]byte{[]byte("min")}, []string{sender.String()}, nil, foreignAssets, suggestedParams, poolAddress, nil, types.Digest{}, [32]byte{}, types.Address{})

	if err != nil {
		return utils.TransactionGroup{}, err
	}

	assetTransferTxn1, err := future.MakeAssetTransferTxn(sender.String(), poolAddress.String(), asset1Amount, nil, suggestedParams, "", asset1ID)

	if err != nil {
		return utils.TransactionGroup{}, err
	}

	var assetTransferTxn2 types.Transaction

	if asset2ID != 0 {
		assetTransferTxn2, err = future.MakeAssetTransferTxn(sender.String(), poolAddress.String(), asset2Amount, nil, suggestedParams, "", asset2ID)
	} else {
		assetTransferTxn2, err = future.MakePaymentTxn(sender.String(), poolAddress.String(), asset2Amount, nil, "", suggestedParams)
	}

	if err != nil {
		return utils.TransactionGroup{}, err
	}

	assetTransferTxn3, err := future.MakeAssetTransferTxn(poolAddress.String(), sender.String(), liquidityAssetAmount, nil, suggestedParams, "", liquidityAssetID)

	if err != nil {
		return utils.TransactionGroup{}, err
	}

	txns := []types.Transaction{paymentTxn, applicationNoOptTxn, assetTransferTxn1, assetTransferTxn2, assetTransferTxn3}

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
