package burn

import (
	"github.com/soheil555/tinyman-mobile-sdk/types"
	"github.com/soheil555/tinyman-mobile-sdk/utils"
	"github.com/soheil555/tinyman-mobile-sdk/v1/contracts"

	"github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/future"
	algoTypes "github.com/algorand/go-algorand-sdk/types"
)

func PrepareBurnTransactions(validatorAppId, asset1ID, asset2ID, liquidityAssetID int, asset1Amount, asset2Amount, liquidityAssetAmount string, senderAddress string, suggestedParams *types.SuggestedParams) (txnGroup *utils.TransactionGroup, err error) {

	sender, err := algoTypes.DecodeAddress(senderAddress)
	if err != nil {
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

	poolLogicsig, err := contracts.GetPoolLogicsig(validatorAppId, asset1ID, asset2ID)

	if err != nil {
		return
	}

	poolAddress := crypto.AddressFromProgram(poolLogicsig.Logic)

	paymentTxn, err := future.MakePaymentTxn(sender.String(), poolAddress.String(), 3000, []byte("fee"), "", algoSuggestedParams)

	if err != nil {
		return
	}

	asset1AmountBig := utils.NewBigIntString(asset1Amount)
	asset2AmountBig := utils.NewBigIntString(asset2Amount)
	liquidityAssetAmountBig := utils.NewBigIntString(liquidityAssetAmount)

	var foreignAssets []uint64

	if asset2ID == 0 {
		foreignAssets = []uint64{uint64(asset1ID), uint64(liquidityAssetID)}
	} else {
		foreignAssets = []uint64{uint64(asset1ID), uint64(asset2ID), uint64(liquidityAssetID)}
	}

	applicationNoOpTxn, err := future.MakeApplicationNoOpTx(uint64(validatorAppId), [][]byte{[]byte("burn")}, []string{sender.String()}, nil, foreignAssets, algoSuggestedParams, poolAddress, nil, algoTypes.Digest{}, [32]byte{}, algoTypes.Address{})

	if err != nil {
		return
	}

	assetTransferTxn1, err := future.MakeAssetTransferTxn(poolAddress.String(), sender.String(), asset1AmountBig.Uint64(), nil, algoSuggestedParams, "", uint64(asset1ID))

	if err != nil {
		return
	}

	var assetTransferTxn2 algoTypes.Transaction

	if asset2ID != 0 {
		assetTransferTxn2, err = future.MakeAssetTransferTxn(poolAddress.String(), sender.String(), asset2AmountBig.Uint64(), nil, algoSuggestedParams, "", uint64(asset2ID))
	} else {
		assetTransferTxn2, err = future.MakePaymentTxn(poolAddress.String(), sender.String(), asset2AmountBig.Uint64(), nil, "", algoSuggestedParams)
	}

	if err != nil {
		return
	}

	assetTransferTxn3, err := future.MakeAssetTransferTxn(sender.String(), poolAddress.String(), liquidityAssetAmountBig.Uint64(), nil, algoSuggestedParams, "", uint64(liquidityAssetID))

	if err != nil {
		return
	}

	txns := []algoTypes.Transaction{paymentTxn, applicationNoOpTxn, assetTransferTxn1, assetTransferTxn2, assetTransferTxn3}

	txnGroup, err = utils.NewTransactionGroup(txns)

	if err != nil {
		return
	}

	lsig := &types.LogicSig{
		Logic: poolLogicsig.Logic,
	}

	err = txnGroup.SignWithLogicsig(lsig)

	return

}
