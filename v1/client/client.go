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
	"github.com/algorand/go-algorand-sdk/client/v2/indexer"
	algoTypes "github.com/algorand/go-algorand-sdk/types"
)

type TinymanClient struct {
	algod          *algod.Client
	indexer        *indexer.Client
	ValidatorAppId int
	assetsCache    map[int]types.Asset
	UserAddress    algoTypes.Address
}

func NewTinymanClient(algodClientURL string, indexerClientURL string, validatorAppId int, userAddress algoTypes.Address) (tinymanClient TinymanClient, err error) {

	headers := []*common.Header{
		{
			Key:   "User-Agent",
			Value: "algosdk",
		},
	}

	algodClient, err := algod.MakeClientWithHeaders(algodClientURL, "", headers)
	if err != nil {
		return
	}

	indexerClient, err := indexer.MakeClientWithHeaders(indexerClientURL, "", headers)
	if err != nil {
		return
	}

	return TinymanClient{
		algodClient,
		indexerClient,
		validatorAppId,
		map[int]types.Asset{},
		userAddress,
	}, nil
}

func NewTinymanTestnetClient(algodClientURL string, indexerClientURL string, userAddress algoTypes.Address) (tinymanClient TinymanClient, err error) {

	return NewTinymanClient(algodClientURL, indexerClientURL, constants.TESTNET_VALIDATOR_APP_ID, userAddress)

}

func NewTinymanMainnetClient(algodClientURL string, indexerClientURL string, userAddress algoTypes.Address) (tinymanClient TinymanClient, err error) {

	return NewTinymanClient(algodClientURL, indexerClientURL, constants.MAINNET_VALIDATOR_APP_ID, userAddress)

}

//TODO: implement later, cycle import error
// func (s *TinymanClient) FetchPool(asset1 interface{}, asset2 interface{}, fetch bool) {
// }

func (s *TinymanClient) FetchAsset(assetID int) (asset types.Asset, err error) {

	if _, ok := s.assetsCache[assetID]; !ok {

		asset = types.Asset{Id: assetID}
		err = asset.Fetch(s.indexer)

		if err != nil {
			return
		}

		s.assetsCache[assetID] = asset

	}

	asset = s.assetsCache[assetID]
	return

}

func (s *TinymanClient) Submit(transactionGroup utils.TransactionGroup, wait bool) (trxInfo models.PendingTransactionInfoResponse, Txid string, err error) {

	signedGroup := transactionGroup.GetSignedGroup()

	sendRawTransaction := s.algod.SendRawTransaction(signedGroup)
	Txid, err = sendRawTransaction.Do(context.Background())

	if err != nil {
		return
	}

	if wait {
		return utils.WaitForConfirmation(s.algod, Txid)
	}

	return

}

func (s *TinymanClient) PrepareAppOptinTransactions(userAddress algoTypes.Address) (txnGroup utils.TransactionGroup, err error) {

	if userAddress.IsZero() {
		userAddress = s.UserAddress
	}

	suggestedParams, err := s.algod.SuggestedParams().Do(context.Background())
	if err != nil {
		return
	}

	txnGroup, err = optin.PrepareAppOptinTransactions(s.ValidatorAppId, userAddress, suggestedParams)

	return

}

func (s *TinymanClient) PrepareAssetOptinTransactions(assetID uint64, userAddress algoTypes.Address) (txnGroup utils.TransactionGroup, err error) {

	if userAddress.IsZero() {
		userAddress = s.UserAddress
	}

	suggestedParams, err := s.algod.SuggestedParams().Do(context.Background())
	if err != nil {
		return
	}

	txnGroup, err = optin.PrepareAssetOptinTransactions(assetID, userAddress, suggestedParams)

	return

}

func (s *TinymanClient) FetchExcessAmounts(userAddress algoTypes.Address) (pools map[string]map[types.Asset]types.AssetAmount, err error) {

	pools = make(map[string]map[types.Asset]types.AssetAmount)

	if userAddress.IsZero() {
		userAddress = s.UserAddress
	}

	_, accountInfo, err := s.indexer.LookupAccountByID(userAddress.String()).Do(context.Background())
	if err != nil {
		return
	}

	var validatorApp models.ApplicationLocalState

	for _, a := range accountInfo.AppsLocalState {

		if a.Id == uint64(s.ValidatorAppId) {
			validatorApp = a
		}

	}

	if reflect.ValueOf(validatorApp).IsZero() {
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

		if bLen >= 9 && b[bLen-9] == byte('e') {

			value := int(validatorAppState[key].Uint)
			var poolAddress string
			poolAddress, err = algoTypes.EncodeAddress(b[:bLen-9])

			if err != nil {
				return
			}

			if pool, ok := pools[poolAddress]; ok {
				pools[poolAddress] = pool
			}

			assetID := binary.BigEndian.Uint64(b[bLen-8:])
			var asset types.Asset
			asset, err = s.FetchAsset(int(assetID))
			if err != nil {
				return
			}

			if pools[poolAddress] == nil {
				pools[poolAddress] = make(map[types.Asset]types.AssetAmount)
			}

			pools[poolAddress][asset] = types.AssetAmount{Asset: asset, Amount: value}

		}

	}

	return

}

func (s *TinymanClient) IsOptedIn(userAddress algoTypes.Address) (bool, error) {

	if userAddress.IsZero() {
		userAddress = s.UserAddress
	}

	_, accountInfo, err := s.indexer.LookupAccountByID(userAddress.String()).Do(context.Background())
	if err != nil {
		return false, err
	}

	for _, a := range accountInfo.AppsLocalState {
		if a.Id == uint64(s.ValidatorAppId) {
			return true, nil
		}
	}

	return false, nil
}

func (s *TinymanClient) AssetIsOptedIn(assetID int, userAddress algoTypes.Address) (bool, error) {

	if userAddress.IsZero() {
		userAddress = s.UserAddress
	}

	_, accountInfo, err := s.indexer.LookupAccountByID(userAddress.String()).Do(context.Background())
	if err != nil {
		return false, err
	}

	for _, a := range accountInfo.Assets {
		if a.AssetId == uint64(assetID) {
			return true, nil
		}
	}

	return false, nil

}
