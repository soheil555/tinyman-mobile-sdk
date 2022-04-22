package burn

import (
	"math/big"
	"tinyman-mobile-sdk/types"
	"tinyman-mobile-sdk/utils"
	"tinyman-mobile-sdk/v1/contracts"

	"github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/future"
	algoTypes "github.com/algorand/go-algorand-sdk/types"
)

func PrepareBurnTransactions(validatorAppId int, asset1ID int, asset2ID int, liquidityAssetID int, asset1Amount string, asset2Amount string, liquidityAssetAmount string, sender []byte, suggestedParams types.SuggestedParams) (txnGroup utils.TransactionGroup, err error) {

	var senderAddress algoTypes.Address
	copy(senderAddress[:], sender)

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

	poolLogicsig, err := contracts.GetPoolLogicsig(validatorAppId, asset1ID, asset2ID)

	if err != nil {
		return
	}

	poolAddress := crypto.AddressFromProgram(poolLogicsig.Logic)

	paymentTxn, err := future.MakePaymentTxn(senderAddress.String(), poolAddress.String(), 3000, []byte("fee"), "", algoSuggestedParams)

	if err != nil {
		return
	}

	//TODO: maybe return error if not ok
	Asset1Amount, ok := new(big.Int).SetString(asset1Amount, 10)
	if !ok {
		return
	}

	Asset2Amount, ok := new(big.Int).SetString(asset2Amount, 10)
	if !ok {
		return
	}

	LiquidityAssetAmount, ok := new(big.Int).SetString(liquidityAssetAmount, 10)
	if !ok {
		return
	}

	var foreignAssets []uint64

	if asset2ID == 0 {
		foreignAssets = []uint64{uint64(asset1ID), uint64(liquidityAssetID)}
	} else {
		foreignAssets = []uint64{uint64(asset1ID), uint64(asset2ID), uint64(liquidityAssetID)}
	}

	applicationNoOpTxn, err := future.MakeApplicationNoOpTx(uint64(validatorAppId), [][]byte{[]byte("burn")}, []string{senderAddress.String()}, nil, foreignAssets, algoSuggestedParams, poolAddress, nil, algoTypes.Digest{}, [32]byte{}, algoTypes.Address{})

	if err != nil {
		return
	}

	assetTransferTxn1, err := future.MakeAssetTransferTxn(poolAddress.String(), senderAddress.String(), Asset1Amount.Uint64(), nil, algoSuggestedParams, "", uint64(asset1ID))

	if err != nil {
		return
	}

	var assetTransferTxn2 algoTypes.Transaction

	if asset2ID != 0 {
		assetTransferTxn2, err = future.MakeAssetTransferTxn(poolAddress.String(), senderAddress.String(), Asset2Amount.Uint64(), nil, algoSuggestedParams, "", uint64(asset2ID))
	} else {
		assetTransferTxn2, err = future.MakePaymentTxn(poolAddress.String(), senderAddress.String(), Asset2Amount.Uint64(), nil, "", algoSuggestedParams)
	}

	if err != nil {
		return
	}

	assetTransferTxn3, err := future.MakeAssetTransferTxn(senderAddress.String(), poolAddress.String(), LiquidityAssetAmount.Uint64(), nil, algoSuggestedParams, "", uint64(liquidityAssetID))

	if err != nil {
		return
	}

	txns := []algoTypes.Transaction{paymentTxn, applicationNoOpTxn, assetTransferTxn1, assetTransferTxn2, assetTransferTxn3}

	txnGroup, err = utils.NewTransactionGroup(txns)

	if err != nil {
		return
	}

	err = txnGroup.SignWithLogicsig(poolLogicsig)

	return

}
