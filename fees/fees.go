package fees

import (
	"crypto/ed25519"
	"fmt"
	"tinyman-mobile-sdk/contracts"
	"tinyman-mobile-sdk/utils"

	"github.com/algorand/go-algorand-sdk/future"
	"github.com/algorand/go-algorand-sdk/logic"
	"github.com/algorand/go-algorand-sdk/types"
)

func PrepareRedeemFeesTransactions(validatorAppId uint64, asset1ID uint64, asset2ID uint64, liquidityAssetID uint64, amount uint64, creator types.Address, sender types.Address, suggestedParams types.SuggestedParams) (utils.TransactionGroup, error) {

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

	applicationNoOptTxn, err := future.MakeApplicationNoOpTx(validatorAppId, [][]byte{[]byte("fees")}, nil, nil, foreignAssets, suggestedParams, poolAddress, nil, types.Digest{}, [32]byte{}, types.Address{})

	if err != nil {
		return utils.TransactionGroup{}, err
	}

	assetTransferTxn, err := future.MakeAssetTransferTxn(poolAddress.String(), creator.String(), amount, nil, suggestedParams, "", liquidityAssetID)

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