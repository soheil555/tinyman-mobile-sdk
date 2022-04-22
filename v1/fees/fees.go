package fees

import (
	"math/big"
	"tinyman-mobile-sdk/types"
	"tinyman-mobile-sdk/utils"
	"tinyman-mobile-sdk/v1/contracts"

	"github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/future"
	algoTypes "github.com/algorand/go-algorand-sdk/types"
)

func PrepareRedeemFeesTransactions(validatorAppId, asset1ID, asset2ID, liquidityAssetID int, amount string, creator, sender []byte, suggestedParams types.SuggestedParams) (txnGroup utils.TransactionGroup, err error) {

	var senderAddress, creatorAddress algoTypes.Address

	copy(senderAddress[:], sender)
	copy(creatorAddress[:], creator)

	Amount, ok := new(big.Int).SetString(amount, 10)
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

	poolLogicsig, err := contracts.GetPoolLogicsig(validatorAppId, asset1ID, asset2ID)

	if err != nil {
		return
	}

	poolAddress := crypto.AddressFromProgram(poolLogicsig.Logic)

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

	applicationNoOpTxn, err := future.MakeApplicationNoOpTx(uint64(validatorAppId), [][]byte{[]byte("fees")}, nil, nil, foreignAssets, algoSuggestedParams, poolAddress, nil, algoTypes.Digest{}, [32]byte{}, algoTypes.Address{})

	if err != nil {
		return
	}

	assetTransferTxn, err := future.MakeAssetTransferTxn(poolAddress.String(), creatorAddress.String(), Amount.Uint64(), nil, algoSuggestedParams, "", uint64(liquidityAssetID))

	if err != nil {
		return
	}

	txns := []algoTypes.Transaction{paymentTxn, applicationNoOpTxn, assetTransferTxn}

	txnGroup, err = utils.NewTransactionGroup(txns)

	if err != nil {
		return
	}

	err = txnGroup.SignWithLogicsig(poolLogicsig)

	return

}
