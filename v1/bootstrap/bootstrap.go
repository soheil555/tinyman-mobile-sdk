package bootstrap

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/soheil555/tinyman-mobile-sdk/types"
	"github.com/soheil555/tinyman-mobile-sdk/utils"
	"github.com/soheil555/tinyman-mobile-sdk/v1/contracts"

	"github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/future"
	algoTypes "github.com/algorand/go-algorand-sdk/types"
)

func hex2Int(hexStr string) uint64 {

	cleaned := strings.Replace(hexStr, "0x", "", -1)

	result, _ := strconv.ParseUint(cleaned, 16, 64)

	return result

}

func PrepareBootstrapTransactions(validatorAppId, asset1ID, asset2ID int, asset1UnitName, asset2UnitName string, senderAddress string, suggestedParams *types.SuggestedParams) (txnGroup *utils.TransactionGroup, err error) {

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

	if asset1ID <= asset2ID {
		err = fmt.Errorf("asset1ID must be greater than to asset2ID")
		return
	}

	if asset2ID == 0 {
		asset2UnitName = "ALGO"
	}

	var paymentTxnAmount uint64
	if asset2ID > 0 {
		paymentTxnAmount = 961000
	} else {
		paymentTxnAmount = 860000
	}
	paymentTxn, err := future.MakePaymentTxn(sender.String(), poolAddress.String(), paymentTxnAmount, []byte("fee"), "", algoSuggestedParams)

	if err != nil {
		return
	}

	var foreignAssets []uint64

	if asset2ID == 0 {
		foreignAssets = []uint64{uint64(asset1ID)}
	} else {
		foreignAssets = []uint64{uint64(asset1ID), uint64(asset2ID)}
	}

	applicationOptInTxn, err := future.MakeApplicationOptInTx(uint64(validatorAppId), [][]byte{[]byte("bootstrap"), utils.IntToBytes(asset1ID), utils.IntToBytes(asset2ID)}, nil, nil, foreignAssets, algoSuggestedParams, poolAddress, nil, algoTypes.Digest{}, [32]byte{}, algoTypes.Address{})

	if err != nil {
		return
	}

	assetCreateTxn, err := future.MakeAssetCreateTxn(poolAddress.String(), nil, algoSuggestedParams, hex2Int("0xFFFFFFFFFFFFFFFF"), 6, false, "", "", "", "", "TMPOOL11", fmt.Sprintf("TinymanPool1.1 {%s}-{%s}", asset1UnitName, asset2UnitName), "https://tinyman.org", "")

	if err != nil {
		return
	}

	assetOptInTxn1, err := future.MakeAssetTransferTxn(poolAddress.String(), poolAddress.String(), 0, nil, algoSuggestedParams, "", uint64(asset1ID))
	if err != nil {
		return
	}

	txns := []algoTypes.Transaction{paymentTxn, applicationOptInTxn, assetCreateTxn, assetOptInTxn1}

	if asset2ID > 0 {

		var assetOptInTxn2 algoTypes.Transaction
		assetOptInTxn2, err = future.MakeAssetTransferTxn(poolAddress.String(), poolAddress.String(), 0, nil, algoSuggestedParams, "", uint64(asset2ID))
		if err != nil {
			return
		}

		txns = append(txns, assetOptInTxn2)

	}

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
