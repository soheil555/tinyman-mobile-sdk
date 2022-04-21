package bootstrap

import (
	"fmt"
	"strconv"
	"strings"
	"tinyman-mobile-sdk/utils"
	"tinyman-mobile-sdk/v1/contracts"

	"github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/future"
	algoTypes "github.com/algorand/go-algorand-sdk/types"
)

func hex2Int(hexStr string) uint64 {

	cleaned := strings.Replace(hexStr, "0x", "", -1)

	result, _ := strconv.ParseUint(cleaned, 16, 64)

	return result

}

func PrepareBootstrapTransactions(validatorAppId uint64, asset1ID uint64, asset2ID uint64, asset1UnitName string, asset2UnitName string, sender algoTypes.Address, suggestedParams algoTypes.SuggestedParams) (txnGroup utils.TransactionGroup, err error) {

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
	paymentTxn, err := future.MakePaymentTxn(sender.String(), poolAddress.String(), paymentTxnAmount, []byte("fee"), "", suggestedParams)

	if err != nil {
		return
	}

	var foreignAssets []uint64

	if asset2ID == 0 {
		foreignAssets = []uint64{asset1ID}
	} else {
		foreignAssets = []uint64{asset1ID, asset2ID}
	}

	applicationOptInTxn, err := future.MakeApplicationOptInTx(validatorAppId, [][]byte{[]byte("bootstrap"), utils.IntToBytes(asset1ID), utils.IntToBytes(asset2ID)}, nil, nil, foreignAssets, suggestedParams, poolAddress, nil, algoTypes.Digest{}, [32]byte{}, algoTypes.Address{})

	if err != nil {
		return
	}

	assetCreateTxn, err := future.MakeAssetCreateTxn(poolAddress.String(), nil, suggestedParams, hex2Int("0xFFFFFFFFFFFFFFFF"), 6, false, "", "", "", "", "TMPOOL11", fmt.Sprintf("TinymanPool1.1 {%s}-{%s}", asset1UnitName, asset2UnitName), "https://tinyman.org", "")

	if err != nil {
		return
	}

	assetOptInTxn1, err := future.MakeAssetTransferTxn(poolAddress.String(), poolAddress.String(), 0, nil, suggestedParams, "", asset1ID)
	if err != nil {
		return
	}

	txns := []algoTypes.Transaction{paymentTxn, applicationOptInTxn, assetCreateTxn, assetOptInTxn1}

	if asset2ID > 0 {

		var assetOptInTxn2 algoTypes.Transaction
		assetOptInTxn2, err = future.MakeAssetTransferTxn(poolAddress.String(), poolAddress.String(), 0, nil, suggestedParams, "", asset2ID)
		if err != nil {
			return
		}

		txns = append(txns, assetOptInTxn2)

	}

	txnGroup, err = utils.NewTransactionGroup(txns)
	if err != nil {
		return
	}

	err = txnGroup.SignWithLogicsig(poolLogicsig)

	return
}
