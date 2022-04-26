package client

import (
	"context"
	b64 "encoding/base64"
	"encoding/binary"
	"encoding/json"
	"math/big"
	"reflect"

	"github.com/soheil555/tinyman-mobile-sdk/types"
	"github.com/soheil555/tinyman-mobile-sdk/utils"
	"github.com/soheil555/tinyman-mobile-sdk/v1/constants"
	"github.com/soheil555/tinyman-mobile-sdk/v1/optin"

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
	assetsCache    map[int]*types.Asset
	UserAddress    string
}

func NewTinymanClient(algodClientURL, indexerClientURL string, validatorAppId int, userAddress string) (tinymanClient *TinymanClient, err error) {

	user, err := algoTypes.DecodeAddress(userAddress)
	if err != nil {
		return
	}

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

	return &TinymanClient{
		algodClient,
		indexerClient,
		validatorAppId,
		map[int]*types.Asset{},
		user.String(),
	}, nil
}

func NewTinymanTestnetClient(algodClientURL, indexerClientURL, userAddress string) (tinymanClient *TinymanClient, err error) {

	return NewTinymanClient(algodClientURL, indexerClientURL, constants.TESTNET_VALIDATOR_APP_ID, userAddress)

}

func NewTinymanMainnetClient(algodClientURL, indexerClientURL, userAddress string) (tinymanClient *TinymanClient, err error) {

	return NewTinymanClient(algodClientURL, indexerClientURL, constants.MAINNET_VALIDATOR_APP_ID, userAddress)

}

//TODO: implement later, cycle import error
// func (s *TinymanClient) FetchPool(asset1 interface{}, asset2 interface{}, fetch bool) {
// }

func (s *TinymanClient) FetchAsset(assetID int) (asset *types.Asset, err error) {

	if _, ok := s.assetsCache[assetID]; !ok {

		asset = &types.Asset{Id: assetID}
		err = asset.Fetch(s.indexer)

		if err != nil {
			return
		}

		s.assetsCache[assetID] = asset

	}

	asset = s.assetsCache[assetID]
	return

}

// not compatible with go-mobile
func (s *TinymanClient) LookupAccountByID(address string) (validRound uint64, result models.Account, err error) {

	return s.indexer.LookupAccountByID(address).Do(context.Background())

}

// not compatible with go-mobile
func (s *TinymanClient) AccountInformation(address string) (response models.Account, err error) {
	return s.algod.AccountInformation(address).Do(context.Background())
}

// not compatible with go-mobile
func (s *TinymanClient) SuggestedParams() (params algoTypes.SuggestedParams, err error) {
	return s.algod.SuggestedParams().Do(context.Background())
}

func (s *TinymanClient) Submit(transactionGroup *utils.TransactionGroup, wait bool) (transactionInformation *types.TransactionInformation, err error) {

	signedGroup := transactionGroup.GetSignedGroup()

	sendRawTransaction := s.algod.SendRawTransaction(signedGroup)
	txid, err := sendRawTransaction.Do(context.Background())

	if err != nil {
		return
	}

	if wait {
		return utils.WaitForConfirmation(s.algod, txid)
	}

	transactionInformation = &types.TransactionInformation{
		TxId: txid,
	}
	return

}

func (s *TinymanClient) PrepareAppOptinTransactions(userAddress string) (txnGroup *utils.TransactionGroup, err error) {

	if len(userAddress) == 0 {
		userAddress = s.UserAddress
	}

	user, err := algoTypes.DecodeAddress(userAddress)
	if err != nil {
		return
	}

	algoSuggestedParams, err := s.algod.SuggestedParams().Do(context.Background())
	if err != nil {
		return
	}

	suggestedParams := &types.SuggestedParams{
		Fee:              int(algoSuggestedParams.Fee),
		GenesisID:        algoSuggestedParams.GenesisID,
		GenesisHash:      algoSuggestedParams.GenesisHash,
		FirstRoundValid:  int(algoSuggestedParams.FirstRoundValid),
		LastRoundValid:   int(algoSuggestedParams.LastRoundValid),
		ConsensusVersion: algoSuggestedParams.ConsensusVersion,
		FlatFee:          algoSuggestedParams.FlatFee,
		MinFee:           int(algoSuggestedParams.MinFee),
	}

	txnGroup, err = optin.PrepareAppOptinTransactions(s.ValidatorAppId, user.String(), suggestedParams)

	return

}

func (s *TinymanClient) PrepareAssetOptinTransactions(assetID int, userAddress string) (txnGroup *utils.TransactionGroup, err error) {

	if len(userAddress) == 0 {
		userAddress = s.UserAddress
	}

	user, err := algoTypes.DecodeAddress(userAddress)
	if err != nil {
		return
	}

	algoSuggestedParams, err := s.algod.SuggestedParams().Do(context.Background())
	if err != nil {
		return
	}

	suggestedParams := &types.SuggestedParams{
		Fee:              int(algoSuggestedParams.Fee),
		GenesisID:        algoSuggestedParams.GenesisID,
		GenesisHash:      algoSuggestedParams.GenesisHash,
		FirstRoundValid:  int(algoSuggestedParams.FirstRoundValid),
		LastRoundValid:   int(algoSuggestedParams.LastRoundValid),
		ConsensusVersion: algoSuggestedParams.ConsensusVersion,
		FlatFee:          algoSuggestedParams.FlatFee,
		MinFee:           int(algoSuggestedParams.MinFee),
	}

	txnGroup, err = optin.PrepareAssetOptinTransactions(assetID, user.String(), suggestedParams)

	return

}

func (s *TinymanClient) FetchExcessAmounts(userAddress string) (excessAmountsStr string, err error) {

	pools := make(map[string]map[int]string)

	if len(userAddress) == 0 {
		userAddress = s.UserAddress
	}

	user, err := algoTypes.DecodeAddress(userAddress)
	if err != nil {
		return
	}

	_, accountInfo, err := s.indexer.LookupAccountByID(user.String()).Do(context.Background())
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

			value := big.NewInt(int64(validatorAppState[key].Uint))
			var poolAddress string
			poolAddress, err = algoTypes.EncodeAddress(b[:bLen-9])

			if err != nil {
				return
			}

			if pool, ok := pools[poolAddress]; ok {
				pools[poolAddress] = pool
			}

			assetID := binary.BigEndian.Uint64(b[bLen-8:])
			var asset *types.Asset
			asset, err = s.FetchAsset(int(assetID))
			if err != nil {
				return
			}

			if pools[poolAddress] == nil {
				pools[poolAddress] = make(map[int]string)
			}

			pools[poolAddress][asset.Id] = value.String()

		}

	}

	excessAmountsBytes, err := json.Marshal(pools)
	if err != nil {
		return
	}
	excessAmountsStr = string(excessAmountsBytes)

	return

}

func (s *TinymanClient) IsOptedIn(userAddress string) (bool, error) {

	if len(userAddress) == 0 {
		userAddress = s.UserAddress
	}

	user, err := algoTypes.DecodeAddress(userAddress)
	if err != nil {
		return false, err
	}

	_, accountInfo, err := s.indexer.LookupAccountByID(user.String()).Do(context.Background())
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

func (s *TinymanClient) AssetIsOptedIn(assetID int, userAddress string) (bool, error) {

	if len(userAddress) == 0 {
		userAddress = s.UserAddress
	}

	user, err := algoTypes.DecodeAddress(userAddress)
	if err != nil {
		return false, err
	}

	_, accountInfo, err := s.indexer.LookupAccountByID(user.String()).Do(context.Background())
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
