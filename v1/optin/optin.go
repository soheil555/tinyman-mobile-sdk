package optin

import (
	"tinyman-mobile-sdk/types"
	"tinyman-mobile-sdk/utils"

	"github.com/algorand/go-algorand-sdk/future"
	algoTypes "github.com/algorand/go-algorand-sdk/types"
)

func PrepareAppOptinTransactions(validatorAppId int, sender []byte, suggestedParams types.SuggestedParams) (txnGroup *utils.TransactionGroup, err error) {

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

	txn, err := future.MakeApplicationOptInTx(uint64(validatorAppId), nil, nil, nil, nil, algoSuggestedParams, senderAddress, nil, algoTypes.Digest{}, [32]byte{}, algoTypes.Address{})

	if err != nil {
		return
	}

	transactions := []algoTypes.Transaction{txn}

	txnGroup, err = utils.NewTransactionGroup(transactions)

	return

}

func PrepareAssetOptinTransactions(assetID int, sender []byte, suggestedParams types.SuggestedParams) (txnGroup *utils.TransactionGroup, err error) {

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

	txn, err := future.MakeAssetTransferTxn(senderAddress.String(), senderAddress.String(), 0, nil, algoSuggestedParams, "", uint64(assetID))
	if err != nil {
		return
	}

	transactions := []algoTypes.Transaction{txn}

	txnGroup, err = utils.NewTransactionGroup(transactions)

	return

}
