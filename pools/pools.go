package pools

import (
	"context"
	"crypto/ed25519"
	b64 "encoding/base64"
	"fmt"
	"tinyman-mobile-sdk/assets"
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
	var pool PoolInfo

	accountInfoResponse, err := accountInfo.Do(context.Background())
	if err != nil {
		return pool, err
	}

	if len(accountInfoResponse.AppsLocalState) == 0 {
		return pool, nil
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
		return pool, err
	}

	//TODO: what is pool address
	_, byteArrays, err := logic.ReadProgram(poolLogicsig.Logic, nil)

	if err != nil {
		return pool, err
	}

	//TODO: where is address in byteArray?
	var poolAddress types.Address

	n := copy(poolAddress[:], byteArrays[1])

	if n != ed25519.PublicKeySize {
		return pool, fmt.Errorf("address generated from receiver bytes is the wrong size")
	}

	if accountInfoResponse.Address != poolAddress.String() {
		return pool, fmt.Errorf("accountInfo address is not equal to poolAddress")
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

	pool = PoolInfo{
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

	return pool, nil

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

func (s *SwapQuote) AmountOutWithSlippage() assets.AssetAmount {

	if s.SwapType == "fixed-output" {
		return s.AmountOut
	}

	s.AmountOut.Mul(uint64(s.Slippage))

}
