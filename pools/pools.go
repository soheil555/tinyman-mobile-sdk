package pools

import (
	"context"
	"crypto/ed25519"
	b64 "encoding/base64"
	"fmt"
	"tinyman-mobile-sdk/assets"
	"tinyman-mobile-sdk/client"
	"tinyman-mobile-sdk/contracts"
	"tinyman-mobile-sdk/utils"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
	"github.com/algorand/go-algorand-sdk/logic"
	"github.com/algorand/go-algorand-sdk/types"
)

//TODO: move to another file
//TODO: Address type
type PoolInfo struct {
	Address                         types.Address
	Asset1ID                        uint64
	Asset2ID                        uint64
	LiquidityAssetID                uint64
	LiquidityAssetName              string
	Asset1Reserves                  uint64
	Asset2Reserves                  uint64
	IssuedLiquidity                 uint64
	UnclaimedProtocolFees           uint64
	OutstandingAsset1Amount         uint64
	OutstandingAsset2Amount         uint64
	OutstandingLiquidityAssetAmount uint64
	ValidatorAppId                  uint64
	AlgoBalance                     uint64
	Round                           uint64
}

func GetPoolInfo(client algod.Client, validatorAppID uint64, asset1ID uint64, asset2ID uint64) (PoolInfo, error) {

	poolLogicsig, err := contracts.GetPoolLogicsig(validatorAppID, asset1ID, asset2ID)
	if err != nil {
		return PoolInfo{}, err
	}

	//TODO: what is pool address
	_, byteArrays, err := logic.ReadProgram(poolLogicsig.Logic, nil)

	if err != nil {
		return PoolInfo{}, err
	}

	//TODO: where is address in byteArray?
	var poolAddress types.Address

	n := copy(poolAddress[:], byteArrays[1])

	if n != ed25519.PublicKeySize {
		return PoolInfo{}, fmt.Errorf("address generated from receiver bytes is the wrong size")
	}

	accountInfo := client.AccountInformation(poolAddress.String())
	return GetPoolInfoFromAccountInfo(accountInfo)

}

func GetPoolInfoFromAccountInfo(accountInfo *algod.AccountInformation) (PoolInfo, error) {

	//TODO: more on make()
	var poolInfo PoolInfo

	accountInfoResponse, err := accountInfo.Do(context.Background())
	if err != nil {
		return poolInfo, err
	}

	if len(accountInfoResponse.AppsLocalState) == 0 {
		return poolInfo, nil
	}

	validatorAppID := accountInfoResponse.AppsLocalState[0].Id

	validatorAppState := make(map[string]models.TealValue)

	for _, x := range accountInfoResponse.AppsLocalState[0].KeyValue {
		validatorAppState[x.Key] = x.Value
	}

	//TODO: fix GetStateInt
	asset1ID := utils.GetStateInt(validatorAppState, "a1")
	asset2ID := utils.GetStateInt(validatorAppState, "a2")

	poolLogicsig, err := contracts.GetPoolLogicsig(validatorAppID, asset1ID, asset2ID)

	if err != nil {
		return poolInfo, err
	}

	//TODO: what is pool address
	_, byteArrays, err := logic.ReadProgram(poolLogicsig.Logic, nil)

	if err != nil {
		return poolInfo, err
	}

	//TODO: where is address in byteArray?
	var poolAddress types.Address

	n := copy(poolAddress[:], byteArrays[1])

	if n != ed25519.PublicKeySize {
		return poolInfo, fmt.Errorf("address generated from receiver bytes is the wrong size")
	}

	if accountInfoResponse.Address != poolAddress.String() {
		return poolInfo, fmt.Errorf("accountInfo address is not equal to poolAddress")
	}

	asset1Reserves := utils.GetStateInt(validatorAppState, "s1")
	asset2Reserves := utils.GetStateInt(validatorAppState, "s2")
	issuedLiquidity := utils.GetStateInt(validatorAppState, "ilt")
	unclaimedProtocolFees := utils.GetStateInt(validatorAppState, "p")

	liquidityAsset := accountInfoResponse.CreatedAssets[0]
	liquidityAssetID := liquidityAsset.Index

	key1 := []byte("o")
	key1 = append(key1, utils.IntToBytes(asset1ID)...)

	b64.StdEncoding.Encode(key1, key1)

	outstandingAsset1Amount := utils.GetStateInt(validatorAppState, key1)

	key2 := []byte("o")
	key2 = append(key2, utils.IntToBytes(asset2ID)...)

	outstandingAsset2Amount := utils.GetStateInt(validatorAppState, key2)

	key3 := []byte("o")
	key3 = append(key3, utils.IntToBytes(liquidityAssetID)...)

	outstandingLiquidityAssetAmount := utils.GetStateInt(validatorAppState, key3)

	poolInfo = PoolInfo{
		Address:                         poolAddress,
		Asset1ID:                        asset1ID,
		Asset2ID:                        asset2ID,
		LiquidityAssetID:                liquidityAsset.Index,
		LiquidityAssetName:              liquidityAsset.Params.Name,
		Asset1Reserves:                  asset1Reserves,
		Asset2Reserves:                  asset2Reserves,
		IssuedLiquidity:                 issuedLiquidity,
		UnclaimedProtocolFees:           unclaimedProtocolFees,
		OutstandingAsset1Amount:         outstandingAsset1Amount,
		OutstandingAsset2Amount:         outstandingAsset2Amount,
		OutstandingLiquidityAssetAmount: outstandingLiquidityAssetAmount,
		ValidatorAppId:                  validatorAppID,
		AlgoBalance:                     accountInfoResponse.Amount,
		Round:                           accountInfoResponse.Round,
	}

	return poolInfo, nil

}

//TODO: maybe all addresses must be string
func GetExcessAssetKey(poolAddress string, assetID uint64) ([]byte, error) {
	a, err := types.DecodeAddress(poolAddress)
	if err != nil {
		return nil, err
	}
	var key []byte
	//TODO: append in one move
	e := []byte("e")
	key = append(key, a[:]...)
	key = append(key, e...)
	key = append(key, utils.IntToBytes(assetID)...)

	return key, nil
}

type SwapQuote struct {
	SwapType  string
	AmountIn  assets.AssetAmount
	AmountOut assets.AssetAmount
	SwapFees  uint64
	Slippage  float64
}

func (s *SwapQuote) AmountOutWithSlippage() (assets.AssetAmount, error) {

	if s.SwapType == "fixed-output" {
		return s.AmountOut, nil
	}

	assetAmount, err := s.AmountOut.Sub(s.AmountOut.Mul(s.Slippage))
	if err != nil {
		return assets.AssetAmount{}, err
	}

	return assetAmount, nil

}

//TODO: pointer or not pointer ?
func (s *SwapQuote) AmountInWithSlippage() (assets.AssetAmount, error) {

	if s.SwapType == "fixed-input" {
		return s.AmountIn, nil
	}

	assetAmount, err := s.AmountIn.Add(s.AmountIn.Mul(s.Slippage))
	if err != nil {
		return assets.AssetAmount{}, err
	}

	return assetAmount, nil

}

func (s *SwapQuote) Price() float64 {
	return s.AmountOut.Amount / s.AmountIn.Amount
}

func (s *SwapQuote) PriceWithSlippage() (float64, error) {

	amountOutWithSlippage, err := s.AmountOutWithSlippage()

	if err != nil {
		return 0, err
	}

	amountInWithSlippage, err := s.AmountInWithSlippage()

	if err != nil {
		return 0, err
	}

	return amountOutWithSlippage.Amount / amountInWithSlippage.Amount, nil

}

type MintQuote struct {
	AmountsIn            map[assets.Asset]assets.AssetAmount
	LiquidityAssetAmount assets.AssetAmount
	Slippage             float64
}

//TODO: in python code it return int?
func (s *MintQuote) LiquidityAssetAmountWithSlippage() (assets.AssetAmount, error) {
	assetAmount, err := s.LiquidityAssetAmount.Sub(s.LiquidityAssetAmount.Mul(s.Slippage))

	if err != nil {
		return assets.AssetAmount{}, err
	}

	return assetAmount, nil
}

type BurnQuote struct {
	AmountsOut           map[assets.Asset]*assets.AssetAmount
	LiquidityAssetAmount assets.AssetAmount
	Slippage             float64
}

func (s *BurnQuote) AmountsOutWithSlippage() (map[assets.Asset]assets.AssetAmount, error) {

	out := make(map[assets.Asset]assets.AssetAmount)

	for k := range s.AmountsOut {
		amountOutWithSlippage, err := s.AmountsOut[k].Sub(s.AmountsOut[k].Mul(s.Slippage))

		if err != nil {
			return nil, err
		}

		out[k] = amountOutWithSlippage
	}

	return out, nil
}

//TODO: is PoolInfo and Pool the same
//TODO: where is address
type Pool struct {
	Client                          client.TinymanClient
	ValidatorAppID                  uint64
	Asset1                          assets.Asset
	Asset2                          assets.Asset
	Exists                          bool
	LiquidityAsset                  assets.Asset
	Asset1Reserves                  uint64
	Asset2Reserves                  uint64
	IssuedLiquidity                 uint64
	UnclaimedProtocolFees           uint64
	OutstandingAsset1Amount         uint64
	OutstandingAsset2Amount         uint64
	OutstandingLiquidityAssetAmount uint64
	LastRefreshedRound              uint64
	AlgoBalance                     uint64
	MinBalance                      uint64
}

//TODO: is validatorID true
func NewPool(client client.TinymanClient, assetA interface{}, assetB interface{}, info interface{}, fetch bool, validatorAppID interface{}) (Pool, error) {

	pool := Pool{}
	pool.Client = client

	if validatorAppID == nil {
		pool.ValidatorAppID = client.ValidatorAppId
	} else {
		validatorAppIDUint, ok := validatorAppID.(uint64)
		if !ok {
			return Pool{}, fmt.Errorf("unsupported type for validatorAppID")
		}
		pool.ValidatorAppID = validatorAppIDUint
	}

	switch v := assetA.(type) {

	case uint64:
		pool.Asset1 = client.FetchAsset(v)
	case assets.Asset:
		pool.Asset1 = v
	default:
		return Pool{}, fmt.Errorf("unsupported type for assetA")

	}

	switch v := assetB.(type) {

	case uint64:
		pool.Asset2 = client.FetchAsset(v)
	case assets.Asset:
		pool.Asset2 = v
	default:
		return Pool{}, fmt.Errorf("unsupported type for assetB")

	}

	if fetch {
		pool.Refresh()
	} else if info != nil {

		switch v := info.(type) {
		case PoolInfo:
			pool.UpdateFromInfo(v)

		default:
			return Pool{}, fmt.Errorf("unsupported type for info")
		}

	}

	return pool, nil

}

func NewPoolFromAccountInfo(accountInfo *algod.AccountInformation, client client.TinymanClient) (Pool, error) {

	info, err := GetPoolInfoFromAccountInfo(accountInfo)

	if err != nil {
		return Pool{}, err
	}

	pool, err := NewPool(client, info.Asset1ID, info.Asset2ID, info, false, info.ValidatorAppId)

	if err != nil {
		return Pool{}, err
	}

	return pool, nil

}

//TODO: is this logic good?
func (s *Pool) RefreshWithInfo(info PoolInfo) {
	s.UpdateFromInfo(info)
}

func (s *Pool) Refresh() {

	info, err := GetPoolInfo(*s.Client.Algod, s.ValidatorAppID, s.Asset1.Id, s.Asset2.Id)
	//TODO:return error maybe
	if err != nil {
		return
	}
	s.UpdateFromInfo(info)

}

//TODO: maybe None value could be -1 or 0
func (s *Pool) UpdateFromInfo(info PoolInfo) {

	//TODO: this is wrong
	if info.LiquidityAssetID != 0 {
		s.Exists = true
	}

	//TODO: asset Id to ID maybe
	s.LiquidityAsset = assets.Asset{Id: info.LiquidityAssetID, Name: info.LiquidityAssetName, UnitName: "TMPOOL11", Decimals: 6}
	s.Asset1Reserves = info.Asset1Reserves
	s.Asset2Reserves = info.Asset2Reserves
	s.IssuedLiquidity = info.IssuedLiquidity
	s.UnclaimedProtocolFees = info.UnclaimedProtocolFees
	s.OutstandingAsset1Amount = info.OutstandingAsset1Amount
	s.OutstandingAsset2Amount = info.OutstandingAsset2Amount
	s.OutstandingLiquidityAssetAmount = info.OutstandingLiquidityAssetAmount
	s.LastRefreshedRound = info.Round

	s.AlgoBalance = info.AlgoBalance
	s.MinBalance = s.GetMinimumBalance()

	if s.Asset2.Id == 0 {
		s.Asset2Reserves = (s.AlgoBalance - s.MinBalance) - s.OutstandingAsset2Amount
	}

}

//TODO: so many uint64 for small numbers
func (s *Pool) GetMinimumBalance() uint64 {

	const (
		MIN_BALANCE_PER_ACCOUNT       uint64 = 100000
		MIN_BALANCE_PER_ASSET         uint64 = 100000
		MIN_BALANCE_PER_APP           uint64 = 100000
		MIN_BALANCE_PER_APP_BYTESLICE uint64 = 50000
		MIN_BALANCE_PER_APP_UINT      uint64 = 28500
	)

	var numAssets uint64
	if s.Asset2.Id == 0 {
		numAssets = 2
	} else {
		numAssets = 3
	}

	var numCreatedApps uint64 = 0
	var numLocalApps uint64 = 1
	var totalUnits uint64 = 16
	var totalByteslices uint64 = 0

	total := MIN_BALANCE_PER_ACCOUNT + (MIN_BALANCE_PER_ASSET * numAssets) + (MIN_BALANCE_PER_APP * (numCreatedApps + numLocalApps)) + MIN_BALANCE_PER_APP_UINT*totalUnits + MIN_BALANCE_PER_APP_BYTESLICE*totalByteslices
	return total
}
