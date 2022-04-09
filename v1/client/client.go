package client

import (
	"context"
	b64 "encoding/base64"
	"encoding/binary"
	"reflect"
	"tinyman-mobile-sdk/types"
	"tinyman-mobile-sdk/utils"
	"tinyman-mobile-sdk/v1/constants"
	"tinyman-mobile-sdk/v1/optin"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/client/v2/common"
	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
	algoTypes "github.com/algorand/go-algorand-sdk/types"
)

type TinymanClient struct {
	Algod          *algod.Client
	ValidatorAppId uint64
	AssetsCache    map[uint64]types.Asset
	UserAddress    algoTypes.Address
}

func NewTinymanClient(algodClient *algod.Client, validatorAppId uint64, userAddress algoTypes.Address) TinymanClient {

	return TinymanClient{
		algodClient,
		validatorAppId,
		map[uint64]types.Asset{},
		userAddress,
	}
}

func NewTinymanTestnetClient(algodClient *algod.Client, userAddress algoTypes.Address) (tinymanClient TinymanClient, err error) {

	//TODO: better way
	//TODO: I think testnet client is changed
	if reflect.DeepEqual(algodClient, algod.Client{}) {

		headers := []*common.Header{
			{Key: "User-Agent", Value: "algosdk"},
		}

		algodClient, err = algod.MakeClientWithHeaders("https://api.testnet.algoexplorer.io", "", headers)

		if err != nil {
			return
		}

	}

	return NewTinymanClient(algodClient, constants.TESTNET_VALIDATOR_APP_ID, userAddress), nil

}

func NewTinymanMainnetClient(algodClient *algod.Client, userAddress algoTypes.Address) (tinymanClient TinymanClient, err error) {

	//TODO: better way
	if reflect.DeepEqual(algodClient, algod.Client{}) {

		headers := []*common.Header{
			{Key: "User-Agent", Value: "algosdk"},
		}

		algodClient, err = algod.MakeClientWithHeaders("https://api.algoexplorer.io", "", headers)

		if err != nil {
			return
		}

	}

	return NewTinymanClient(algodClient, constants.MAINNET_VALIDATOR_APP_ID, userAddress), nil

}

//TODO: implement later
//TODO: cycle import error
func (s *TinymanClient) FetchPool(asset1 interface{}, asset2 interface{}, fetch bool) {
}

func (s *TinymanClient) FetchAsset(assetID uint64) types.Asset {

	if _, ok := s.AssetsCache[assetID]; !ok {

		asset := types.Asset{Id: assetID}
		asset.Fetch(s.Algod)
		s.AssetsCache[assetID] = asset

	}

	return s.AssetsCache[assetID]

}

func (s *TinymanClient) Submit(transactionGroup utils.TransactionGroup, wait bool) (pendingTrxInfo models.PendingTransactionInfoResponse, Txid string, err error) {

	//TODO: maybe better way
	var signedGroup []byte

	for _, txn := range transactionGroup.SignedTransactions {
		signedGroup = append(signedGroup, txn...)
	}

	sendRawTransaction := s.Algod.SendRawTransaction(signedGroup)
	Txid, err = sendRawTransaction.Do(context.Background())

	if err != nil {
		return
	}

	if wait {
		return utils.WaitForConfirmation(s.Algod, Txid)
	}

	return

}

func (s *TinymanClient) PrepareAppOptinTransactions(userAddress algoTypes.Address) (txnGroup utils.TransactionGroup, err error) {

	if (userAddress == algoTypes.Address{}) {
		userAddress = s.UserAddress
	}

	suggestedParams, err := s.Algod.SuggestedParams().Do(context.Background())
	if err != nil {
		return
	}

	txnGroup, err = optin.PrepareAppOptinTransactions(s.ValidatorAppId, userAddress, suggestedParams)

	if err != nil {
		return
	}

	return

}

func (s *TinymanClient) PrepareAssetOptinTransactions(assetID uint64, userAddress algoTypes.Address) (txnGroup utils.TransactionGroup, err error) {

	if (userAddress == algoTypes.Address{}) {
		userAddress = s.UserAddress
	}

	suggestedParams, err := s.Algod.SuggestedParams().Do(context.Background())
	if err != nil {
		return
	}

	txnGroup, err = optin.PrepareAssetOptinTransactions(assetID, userAddress, suggestedParams)

	if err != nil {
		return
	}

	return

}

func (s *TinymanClient) FetchExcessAmounts(userAddress algoTypes.Address) (pools map[string]map[types.Asset]types.AssetAmount, err error) {

	//TODO: is pools type ok?
	pools = make(map[string]map[types.Asset]types.AssetAmount)

	if (userAddress == algoTypes.Address{}) {
		userAddress = s.UserAddress
	}

	accountInfo, err := s.Algod.AccountInformation(userAddress.String()).Do(context.Background())
	if err != nil {
		return
	}

	var validatorApps []models.ApplicationLocalState
	var validatorApp models.ApplicationLocalState

	for _, a := range accountInfo.AppsLocalState {

		if a.Id == s.ValidatorAppId {
			validatorApps = append(validatorApps, a)
		}

	}

	if len(validatorApps) > 0 {
		validatorApp = validatorApps[0]
	} else {
		return
	}

	validatorAppState := make(map[string]models.TealValue)

	for _, x := range validatorApp.KeyValue {

		validatorAppState[x.Key] = x.Value

	}

	for key := range validatorAppState {
		var b []byte
		b, err = b64.StdEncoding.DecodeString(key)
		if err != nil {
			return
		}

		bLen := len(b)

		//TODO: is it correct?
		if bLen >= 9 && b[bLen-9] == 101 {
			value := validatorAppState[key].Uint
			var poolAddress string
			poolAddress, err = algoTypes.EncodeAddress(b[:bLen-9])

			if err != nil {
				return
			}

			if pool, ok := pools[poolAddress]; ok {
				pools[poolAddress] = pool
			}

			assetID := binary.BigEndian.Uint64(b[bLen-8:])
			asset := s.FetchAsset(assetID)
			pools[poolAddress][asset] = types.AssetAmount{Asset: asset, Amount: float64(value)}

		}

	}

	return

}

func (s *TinymanClient) IsOptIn(userAddress algoTypes.Address) (bool, error) {

	if (userAddress == algoTypes.Address{}) {
		userAddress = s.UserAddress
	}

	accountInfo, err := s.Algod.AccountInformation(userAddress.String()).Do(context.Background())
	if err != nil {
		return false, err
	}

	for _, a := range accountInfo.AppsLocalState {
		if a.Id == s.ValidatorAppId {
			return true, nil
		}
	}

	return false, nil
}

func (s *TinymanClient) AssetIsOptedIn(assetID uint64, userAddress algoTypes.Address) (bool, error) {

	if (userAddress == algoTypes.Address{}) {
		userAddress = s.UserAddress
	}

	accountInfo, err := s.Algod.AccountInformation(userAddress.String()).Do(context.Background())
	if err != nil {
		return false, err
	}

	for _, a := range accountInfo.Assets {
		if a.AssetId == assetID {
			return true, nil
		}
	}

	return false, nil

}
