package swap

import (
	"math/big"
	"tinyman-mobile-sdk/types"
	"tinyman-mobile-sdk/utils"
	"tinyman-mobile-sdk/v1/contracts"

	"github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/future"
	algoTypes "github.com/algorand/go-algorand-sdk/types"
)

func PrepareSwapTransactions(validatorAppId, asset1ID, asset2ID, liquidityAssetID, assetInID int, assetInAmount, assetOutAmount string, swapType string, sender []byte, suggestedParams types.SuggestedParams) (txnGroup *utils.TransactionGroup, err error) {

	poolLogicsig, err := contracts.GetPoolLogicsig(validatorAppId, asset1ID, asset2ID)

	if err != nil {
		return
	}

	AssetInAmount, ok := new(big.Int).SetString(assetInAmount, 10)
	if !ok {
		return
	}

	AssetOutAmount, ok := new(big.Int).SetString(assetOutAmount, 10)
	if !ok {
		return
	}

	algoSuggestedParams := algoTypes.SuggestedParams{
		Fee:              algoTypes.MicroAlgos(suggestedParams.Fee),
		GenesisID:        suggestedParams.GenesisID,
		GenesisHash:      suggestedParams.GenesisHash,
		FirstRoundValid:  algoTypes.Round(suggestedParams.FirstRoundValid),
		LastRoundValid:   algoTypes.Round(suggestedParams.LastRoundValid),
		ConsensusVersion: suggestedParams.ConsensusVersion,
		FlatFee:          suggestedParams.FlatFee,
		MinFee:           uint64(suggestedParams.MinFee),
	}

	var senderAddress algoTypes.Address
	copy(senderAddress[:], sender)

	poolAddress := crypto.AddressFromProgram(poolLogicsig.Logic)

	swapTypes := map[string]string{
		"fixed-input":  "fi",
		"fixed-output": "fo",
	}

	var assetOutID int

	if assetInID == asset1ID {
		assetOutID = asset2ID
	} else {
		assetOutID = asset1ID
	}

	paymentTxn, err := future.MakePaymentTxn(senderAddress.String(), poolAddress.String(), 2000, []byte("fee"), "", algoSuggestedParams)

	if err != nil {
		return
	}

	var foreignAssets []uint64

	if asset2ID == 0 {
		foreignAssets = []uint64{uint64(asset1ID), uint64(liquidityAssetID)}
	} else {
		foreignAssets = []uint64{uint64(asset1ID), uint64(asset2ID), uint64(liquidityAssetID)}
	}

	applicationNoOpTxn, err := future.MakeApplicationNoOpTx(uint64(validatorAppId), [][]byte{[]byte("swap"), []byte(swapTypes[swapType])}, []string{senderAddress.String()}, nil, foreignAssets, algoSuggestedParams, poolAddress, nil, algoTypes.Digest{}, [32]byte{}, algoTypes.Address{})

	if err != nil {
		return
	}

	var assetTransferInTxn algoTypes.Transaction

	if assetInID != 0 {
		assetTransferInTxn, err = future.MakeAssetTransferTxn(senderAddress.String(), poolAddress.String(), AssetInAmount.Uint64(), nil, algoSuggestedParams, "", uint64(assetInID))
	} else {
		assetTransferInTxn, err = future.MakePaymentTxn(senderAddress.String(), poolAddress.String(), AssetInAmount.Uint64(), nil, "", algoSuggestedParams)
	}

	if err != nil {
		return
	}

	var assetTransferOutTxn algoTypes.Transaction

	if assetOutID != 0 {
		assetTransferOutTxn, err = future.MakeAssetTransferTxn(poolAddress.String(), senderAddress.String(), AssetOutAmount.Uint64(), nil, algoSuggestedParams, "", uint64(assetOutID))
	} else {
		assetTransferOutTxn, err = future.MakePaymentTxn(poolAddress.String(), senderAddress.String(), AssetOutAmount.Uint64(), nil, "", algoSuggestedParams)
	}

	if err != nil {
		return
	}

	txns := []algoTypes.Transaction{paymentTxn, applicationNoOpTxn, assetTransferInTxn, assetTransferOutTxn}

	txnGroup, err = utils.NewTransactionGroup(txns)

	if err != nil {
		return
	}

	err = txnGroup.SignWithLogicsig(poolLogicsig)

	return

}
