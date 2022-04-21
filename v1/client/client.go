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
	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
	"github.com/algorand/go-algorand-sdk/client/v2/indexer"
	algoTypes "github.com/algorand/go-algorand-sdk/types"
)

type TinymanClient struct {
	Algod          *algod.Client
	Indexer        *indexer.Client
	ValidatorAppId uint64
	AssetsCache    map[uint64]types.Asset
	UserAddress    algoTypes.Address
}

func NewTinymanClient(algodClient *algod.Client, indexerClient *indexer.Client, validatorAppId uint64, userAddress algoTypes.Address) TinymanClient {

	return TinymanClient{
		algodClient,
		indexerClient,
		validatorAppId,
		map[uint64]types.Asset{},
		userAddress,
	}
}

func NewTinymanTestnetClient(algodClient *algod.Client, indexerClient *indexer.Client, userAddress algoTypes.Address) (tinymanClient TinymanClient, err error) {

	return NewTinymanClient(algodClient, indexerClient, constants.TESTNET_VALIDATOR_APP_ID, userAddress), nil

}

func NewTinymanMainnetClient(algodClient *algod.Client, indexerClient *indexer.Client, userAddress algoTypes.Address) (tinymanClient TinymanClient, err error) {

	return NewTinymanClient(algodClient, indexerClient, constants.MAINNET_VALIDATOR_APP_ID, userAddress), nil

}

//TODO: implement later, cycle import error
func (s *TinymanClient) FetchPool(asset1 interface{}, asset2 interface{}, fetch bool) {
}

func (s *TinymanClient) FetchAsset(assetID uint64) (asset types.Asset, err error) {

	if _, ok := s.AssetsCache[assetID]; !ok {

		asset = types.Asset{Id: assetID}
		err = asset.Fetch(s.Indexer)

		if err != nil {
			return
		}

		s.AssetsCache[assetID] = asset

	}

	asset = s.AssetsCache[assetID]
	return

}

func (s *TinymanClient) Submit(transactionGroup utils.TransactionGroup, wait bool) (trxInfo models.PendingTransactionInfoResponse, Txid string, err error) {

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

	if userAddress.IsZero() {
		userAddress = s.UserAddress
	}

	suggestedParams, err := s.Algod.SuggestedParams().Do(context.Background())
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

	suggestedParams, err := s.Algod.SuggestedParams().Do(context.Background())
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

	_, accountInfo, err := s.Indexer.LookupAccountByID(userAddress.String()).Do(context.Background())
	if err != nil {
		return
	}

	var validatorApp models.ApplicationLocalState

	for _, a := range accountInfo.AppsLocalState {

		if a.Id == s.ValidatorAppId {
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
			var asset types.Asset
			asset, err = s.FetchAsset(assetID)
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

	_, accountInfo, err := s.Indexer.LookupAccountByID(userAddress.String()).Do(context.Background())
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

	if userAddress.IsZero() {
		userAddress = s.UserAddress
	}

	_, accountInfo, err := s.Indexer.LookupAccountByID(userAddress.String()).Do(context.Background())
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
