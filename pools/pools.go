package pools

import (
	"context"
	"crypto/ed25519"
	b64 "encoding/base64"
	"fmt"
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

func GetPoolInfo(client algod.Client, validatorAppID uint64, asset1ID uint64, asset2ID uint64) (map[string]interface{}, error) {

	poolLogicsig, err := contracts.GetPoolLogicsig(validatorAppID, asset1ID, asset2ID)
	if err != nil {
		return nil, err
	}

	//TODO: what is pool address
	_, byteArrays, err := logic.ReadProgram(poolLogicsig.Logic, nil)

	if err != nil {
		return nil, err
	}

	//TODO: where is address in byteArray?
	var poolAddress types.Address

	n := copy(poolAddress[:], byteArrays[1])

	if n != ed25519.PublicKeySize {
		return nil, fmt.Errorf("address generated from receiver bytes is the wrong size")
	}

	accountInfo := client.AccountInformation(poolAddress.String())
	return GetPoolInfoFromAccountInfo(accountInfo)

}

func GetPoolInfoFromAccountInfo(accountInfo *algod.AccountInformation) (map[string]interface{}, error) {

	//TODO: more on make()
	var pool map[string]interface{}

	accountInfoResponse, err := accountInfo.Do(context.Background())
	if err != nil {
		return nil, err
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
		return nil, err
	}

	//TODO: what is pool address
	_, byteArrays, err := logic.ReadProgram(poolLogicsig.Logic, nil)

	if err != nil {
		return nil, err
	}

	//TODO: where is address in byteArray?
	var poolAddress types.Address

	n := copy(poolAddress[:], byteArrays[1])

	if n != ed25519.PublicKeySize {
		return nil, fmt.Errorf("address generated from receiver bytes is the wrong size")
	}

	if accountInfoResponse.Address != poolAddress.String() {
		return nil, fmt.Errorf("accountInfo address is not equal to poolAddress")
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

	return pool, nil

}
