package bootstrap

import (
	"crypto/ed25519"
	"fmt"
	"strconv"
	"strings"
	"tinyman-mobile-sdk/contracts"
	"tinyman-mobile-sdk/utils"

	"github.com/algorand/go-algorand-sdk/future"
	"github.com/algorand/go-algorand-sdk/logic"
	"github.com/algorand/go-algorand-sdk/types"
)

func Hex2Int(hexStr string) uint64 {

	cleaned := strings.Replace(hexStr, "0x", "", -1)

	result, _ := strconv.ParseUint(cleaned, 16, 64)

	return result

}

func PrepareBootstrapTransactions(validatorAppId uint64, asset1ID uint64, asset2ID uint64, asset1UnitName string, asset2UnitName string, sender types.Address, suggestedParams types.SuggestedParams) (utils.TransactionGroup, error) {

	poolLogicsig, err := contracts.GetPoolLogicsig(validatorAppId, asset1ID, asset2ID)

	if err != nil {
		return utils.TransactionGroup{}, err
	}

	//TODO: what is pool address
	_, bytesArrays, err := logic.ReadProgram(poolLogicsig.Logic, nil)

	if err != nil {
		return utils.TransactionGroup{}, err
	}

	var poolAddress types.Address

	n := copy(poolAddress[:], bytesArrays[1])

	if n != ed25519.PublicKeySize {
		return utils.TransactionGroup{}, fmt.Errorf("address generated from receiver bytes is the wrong size")
	}

	if asset1ID > asset2ID {
		return utils.TransactionGroup{}, fmt.Errorf("asset2ID must be greate than equal asset1ID")
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
		return utils.TransactionGroup{}, err
	}

	var foreignAssets []uint64

	if asset2ID == 0 {
		foreignAssets = []uint64{asset1ID}
	} else {
		foreignAssets = []uint64{asset1ID, asset2ID}
	}

	applicationOptInTxn, err := future.MakeApplicationOptInTx(validatorAppId, [][]byte{[]byte("bootstrap"), utils.IntToBytes(asset1ID), utils.IntToBytes(asset2ID)}, nil, nil, foreignAssets, suggestedParams, poolAddress, nil, types.Digest{}, [32]byte{}, types.Address{})

	if err != nil {
		return utils.TransactionGroup{}, err
	}

	assetCreateTxn, err := future.MakeAssetCreateTxn(poolAddress.String(), nil, suggestedParams, Hex2Int("0xFFFFFFFFFFFFFFFF"), 6, false, "", "", "", "", "TMPOOL11", fmt.Sprintf("TinymanPool1.1 {%s}-{%s}", asset1UnitName, asset2UnitName), "https://tinyman.org", "")

	if err != nil {
		return utils.TransactionGroup{}, err
	}

	//TODO: is it the same as AssetOptInTxn
	assetOptInTxn1, err := future.MakeAssetTransferTxn(poolAddress.String(), poolAddress.String(), 0, nil, suggestedParams, "", asset1ID)
	if err != nil {
		return utils.TransactionGroup{}, err
	}

	txns := []types.Transaction{paymentTxn, applicationOptInTxn, assetCreateTxn, assetOptInTxn1}

	if asset2ID > 0 {

		assetOptInTxn2, err := future.MakeAssetTransferTxn(poolAddress.String(), poolAddress.String(), 0, nil, suggestedParams, "", asset2ID)
		if err != nil {
			return utils.TransactionGroup{}, err
		}

		txns = append(txns, assetOptInTxn2)

	}

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
