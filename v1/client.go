package client

import (
	"context"
	b64 "encoding/base64"
	"encoding/binary"
	"reflect"
	"tinyman-mobile-sdk/assets"
	"tinyman-mobile-sdk/utils"
	"tinyman-mobile-sdk/v1/constants"
	"tinyman-mobile-sdk/v1/optin"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/client/v2/common"
	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
	"github.com/algorand/go-algorand-sdk/types"
)

type TinymanClient struct {
	Algod          *algod.Client
	ValidatorAppId uint64
	AssetsCache    map[uint64]assets.Asset
	UserAddress    types.Address
}

func NewTinymanClient(algodClient *algod.Client, validatorAppId uint64, userAddress types.Address) TinymanClient {

	return TinymanClient{
		algodClient,
		validatorAppId,
		map[uint64]assets.Asset{},
		userAddress,
	}
}

func NewTinymanTestnetClient(algodClient *algod.Client, userAddress types.Address) (TinymanClient, error) {

	//TODO: better way
	if reflect.DeepEqual(algodClient, algod.Client{}) {

		headers := []*common.Header{
			{Key: "User-Agent", Value: "algosdk"},
		}

		var err error

		algodClient, err = algod.MakeClientWithHeaders("https://api.testnet.algoexplorer.io", "", headers)

		if err != nil {
			return TinymanClient{}, err
		}

	}

	return NewTinymanClient(algodClient, constants.TESTNET_VALIDATOR_APP_ID, userAddress), nil

}

func NewTinymanMainnetClient(algodClient *algod.Client, userAddress types.Address) (TinymanClient, error) {

	//TODO: better way
	if reflect.DeepEqual(algodClient, algod.Client{}) {

		headers := []*common.Header{
			{Key: "User-Agent", Value: "algosdk"},
		}

		var err error
		algodClient, err = algod.MakeClientWithHeaders("https://api.algoexplorer.io", "", headers)

		if err != nil {
			return TinymanClient{}, err
		}

	}

	return NewTinymanClient(algodClient, constants.MAINNET_VALIDATOR_APP_ID, userAddress), nil

}

//TODO: implement later
func (s *TinymanClient) FetchPool(asset1 interface{}, asset2 interface{}, fetch bool) {

}

func (s *TinymanClient) FetchAsset(assetID uint64) assets.Asset {

	if _, ok := s.AssetsCache[assetID]; !ok {

		asset := assets.Asset{Id: assetID}
		asset.Fetch(s.Algod)
		s.AssetsCache[assetID] = asset

	}

	return s.AssetsCache[assetID]

}

func (s *TinymanClient) Submit(transactionGroup utils.TransactionGroup, wait bool) (*models.PendingTransactionInfoResponse, string, error) {

	//TODO: maybe better way
	var signedGroup []byte

	for _, txn := range transactionGroup.SignedTransactions {
		signedGroup = append(signedGroup, txn...)
	}

	sendRawTransaction := s.Algod.SendRawTransaction(signedGroup)
	txid, err := sendRawTransaction.Do(context.Background())

	if err != nil {
		return nil, "", err
	}

	if wait {
		return utils.WaitForConfirmation(s.Algod, txid)
	}

	return nil, txid, nil

}

func (s *TinymanClient) PrepareAppOptinTransactions(userAddress types.Address) (utils.TransactionGroup, error) {

	if (userAddress == types.Address{}) {
		userAddress = s.UserAddress
	}

	suggestedParams, err := s.Algod.SuggestedParams().Do(context.Background())
	if err != nil {
		return utils.TransactionGroup{}, err
	}

	txnGroup, err := optin.PrepareAppOptinTransactions(s.ValidatorAppId, userAddress, suggestedParams)

	if err != nil {
		return utils.TransactionGroup{}, err
	}

	return txnGroup, nil

}

func (s *TinymanClient) PrepareAssetOptinTransactions(assetID uint64, userAddress types.Address) (utils.TransactionGroup, error) {

	if (userAddress == types.Address{}) {
		userAddress = s.UserAddress
	}

	suggestedParams, err := s.Algod.SuggestedParams().Do(context.Background())
	if err != nil {
		return utils.TransactionGroup{}, err
	}

	txnGroup, err := optin.PrepareAssetOptinTransactions(assetID, userAddress, suggestedParams)

	if err != nil {
		return utils.TransactionGroup{}, err
	}

	return txnGroup, nil

}

func (s *TinymanClient) FetchExcessAmounts(userAddress types.Address) (map[string]map[assets.Asset]assets.AssetAmount, error) {

	//TODO: is pools type ok?
	pools := make(map[string]map[assets.Asset]assets.AssetAmount)

	if (userAddress == types.Address{}) {
		userAddress = s.UserAddress
	}

	accountInfo, err := s.Algod.AccountInformation(userAddress.String()).Do(context.Background())
	if err != nil {
		return nil, err
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
		return pools, nil
	}

	validatorAppState := make(map[string]models.TealValue)

	for _, x := range validatorApp.KeyValue {

		validatorAppState[x.Key] = x.Value

	}

	for key := range validatorAppState {
		b, err := b64.StdEncoding.DecodeString(key)
		if err != nil {
			return nil, err
		}

		bLen := len(b)

		//TODO: is it correct?
		if bLen >= 9 && b[bLen-9] == 101 {
			value := validatorAppState[key].Uint
			poolAddress, err := types.EncodeAddress(b[:bLen-9])

			if err != nil {
				return nil, err
			}

			if pool, ok := pools[poolAddress]; ok {
				pools[poolAddress] = pool
			}

			assetID := binary.BigEndian.Uint64(b[bLen-8:])
			asset := s.FetchAsset(assetID)
			pools[poolAddress][asset] = assets.AssetAmount{Asset: asset, Amount: float64(value)}

		}

	}

	return pools, nil

}

func (s *TinymanClient) IsOptIn(userAddress types.Address) (bool, error) {

	if (userAddress == types.Address{}) {
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

func (s *TinymanClient) AssetIsOptedIn(assetID uint64, userAddress types.Address) (bool, error) {

	if (userAddress == types.Address{}) {
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
