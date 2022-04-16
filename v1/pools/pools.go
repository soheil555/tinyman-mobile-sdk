package pools

import (
	"context"
	b64 "encoding/base64"
	"fmt"
	"math"
	"tinyman-mobile-sdk/types"
	"tinyman-mobile-sdk/utils"
	"tinyman-mobile-sdk/v1/bootstrap"
	"tinyman-mobile-sdk/v1/burn"
	"tinyman-mobile-sdk/v1/client"
	"tinyman-mobile-sdk/v1/contracts"
	"tinyman-mobile-sdk/v1/fees"
	"tinyman-mobile-sdk/v1/mint"
	"tinyman-mobile-sdk/v1/optin"
	"tinyman-mobile-sdk/v1/redeem"
	"tinyman-mobile-sdk/v1/swap"

	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
	"github.com/algorand/go-algorand-sdk/client/v2/indexer"
	"github.com/algorand/go-algorand-sdk/crypto"
	algoTypes "github.com/algorand/go-algorand-sdk/types"
)

//TODO: move to another file
//TODO: Address type
//TODO: round vs lastRefreshedRound
type PoolInfo struct {
	Address                         algoTypes.Address
	Asset1ID                        uint64
	Asset2ID                        uint64
	Asset1UnitName                  string
	Asset2UnitName                  string
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
	LastRefreshedRound              uint64
}

func GetPoolInfo(indexer *indexer.Client, validatorAppID uint64, asset1ID uint64, asset2ID uint64) (poolInfo PoolInfo, err error) {

	poolLogicsig, err := contracts.GetPoolLogicsig(validatorAppID, asset1ID, asset2ID)
	if err != nil {
		return
	}

	poolAddress := crypto.AddressFromProgram(poolLogicsig.Logic)

	accountInfo := indexer.LookupAccountByID(poolAddress.String())
	return GetPoolInfoFromAccountInfo(accountInfo)

}

func GetPoolInfoFromAccountInfo(accountInfo *indexer.LookupAccountByID) (poolInfo PoolInfo, err error) {

	//TODO: more on make()

	_, accountInfoResponse, err := accountInfo.Do(context.Background())
	if err != nil {
		return
	}

	if len(accountInfoResponse.AppsLocalState) == 0 {
		return
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
		return
	}

	poolAddress := crypto.AddressFromProgram(poolLogicsig.Logic)

	if accountInfoResponse.Address != poolAddress.String() {
		err = fmt.Errorf("accountInfo address is not equal to poolAddress")
		return
	}

	asset1Reserves := utils.GetStateInt(validatorAppState, "s1")
	asset2Reserves := utils.GetStateInt(validatorAppState, "s2")
	issuedLiquidity := utils.GetStateInt(validatorAppState, "ilt")
	unclaimedProtocolFees := utils.GetStateInt(validatorAppState, "p")

	liquidityAsset := accountInfoResponse.CreatedAssets[0]
	liquidityAssetID := liquidityAsset.Index

	key1 := []byte("o")
	key1 = append(key1, utils.IntToBytes(asset1ID)...)
	encodedKey1 := make([]byte, b64.StdEncoding.EncodedLen(len(key1)))

	b64.StdEncoding.Encode(encodedKey1, key1)

	outstandingAsset1Amount := utils.GetStateInt(validatorAppState, encodedKey1)

	key2 := []byte("o")
	key2 = append(key2, utils.IntToBytes(asset2ID)...)
	encodedKey2 := make([]byte, b64.StdEncoding.EncodedLen(len(key2)))

	outstandingAsset2Amount := utils.GetStateInt(validatorAppState, encodedKey2)

	key3 := []byte("o")
	key3 = append(key3, utils.IntToBytes(liquidityAssetID)...)
	encodedKey3 := make([]byte, b64.StdEncoding.EncodedLen(len(key3)))

	outstandingLiquidityAssetAmount := utils.GetStateInt(validatorAppState, encodedKey3)

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

	return

}

//TODO: maybe all addresses must be string
func GetExcessAssetKey(poolAddress string, assetID uint64) (key []byte, err error) {
	a, err := algoTypes.DecodeAddress(poolAddress)
	if err != nil {
		return
	}
	//TODO: append in one move
	e := []byte("e")
	key = append(key, a[:]...)
	key = append(key, e...)
	key = append(key, utils.IntToBytes(assetID)...)

	return
}

type SwapQuote struct {
	SwapType  string
	AmountIn  types.AssetAmount
	AmountOut types.AssetAmount
	SwapFees  types.AssetAmount
	Slippage  float64
}

func (s *SwapQuote) AmountOutWithSlippage() (assetAmount types.AssetAmount, err error) {

	if s.SwapType == "fixed-output" {
		return s.AmountOut, nil
	}

	assetAmount, err = s.AmountOut.Sub(s.AmountOut.Mul(s.Slippage))
	if err != nil {
		return
	}

	return

}

//TODO: pointer or not pointer ?
func (s *SwapQuote) AmountInWithSlippage() (assetAmount types.AssetAmount, err error) {

	if s.SwapType == "fixed-input" {
		return s.AmountIn, nil
	}

	assetAmount, err = s.AmountIn.Add(s.AmountIn.Mul(s.Slippage))
	if err != nil {
		return
	}

	return

}

func (s *SwapQuote) Price() uint64 {
	return s.AmountOut.Amount / s.AmountIn.Amount
}

func (s *SwapQuote) PriceWithSlippage() (uint64, error) {

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
	AmountsIn            map[types.Asset]types.AssetAmount
	LiquidityAssetAmount types.AssetAmount
	Slippage             float64
}

//TODO: in python code it return int?
func (s *MintQuote) LiquidityAssetAmountWithSlippage() (assetAmount types.AssetAmount, err error) {
	assetAmount, err = s.LiquidityAssetAmount.Sub(s.LiquidityAssetAmount.Mul(s.Slippage))

	//TODO: it is not required
	if err != nil {
		return
	}

	return
}

type BurnQuote struct {
	AmountsOut           map[types.Asset]*types.AssetAmount
	LiquidityAssetAmount types.AssetAmount
	Slippage             float64
}

func (s *BurnQuote) AmountsOutWithSlippage() (out map[types.Asset]types.AssetAmount, err error) {

	out = make(map[types.Asset]types.AssetAmount)

	for k := range s.AmountsOut {
		var amountOutWithSlippage types.AssetAmount
		amountOutWithSlippage, err = s.AmountsOut[k].Sub(s.AmountsOut[k].Mul(s.Slippage))

		if err != nil {
			return
		}

		out[k] = amountOutWithSlippage
	}

	return
}

//TODO: is PoolInfo and Pool the same
//TODO: where is address
type Pool struct {
	Client                          client.TinymanClient
	ValidatorAppID                  uint64
	Asset1                          types.Asset
	Asset2                          types.Asset
	Exists                          bool
	LiquidityAsset                  types.Asset
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
func MakePool(client client.TinymanClient, assetA interface{}, assetB interface{}, info interface{}, fetch bool, validatorAppID interface{}) (pool Pool, err error) {

	pool.Client = client

	if validatorAppID == nil {
		pool.ValidatorAppID = client.ValidatorAppId
	} else {
		validatorAppIDUint, ok := validatorAppID.(uint64)
		if !ok {
			err = fmt.Errorf("unsupported type for validatorAppID")
			return
		}
		pool.ValidatorAppID = validatorAppIDUint
	}

	switch v := assetA.(type) {

	case uint64:
		pool.Asset1 = client.FetchAsset(v)
	case types.Asset:
		pool.Asset1 = v
	default:
		err = fmt.Errorf("unsupported type for assetA")
		return

	}

	switch v := assetB.(type) {

	case uint64:
		pool.Asset2 = client.FetchAsset(v)
	case types.Asset:
		pool.Asset2 = v
	default:
		err = fmt.Errorf("unsupported type for assetB")
		return

	}

	if fetch {
		err = pool.Refresh()
		if err != nil {
			return
		}
	} else if info != nil {

		switch v := info.(type) {
		case PoolInfo:
			pool.UpdateFromInfo(v)

		default:
			err = fmt.Errorf("unsupported type for info")
			return
		}

	}

	return

}

func MakePoolFromAccountInfo(accountInfo *indexer.LookupAccountByID, client client.TinymanClient) (pool Pool, err error) {

	info, err := GetPoolInfoFromAccountInfo(accountInfo)

	if err != nil {
		return
	}

	pool, err = MakePool(client, info.Asset1ID, info.Asset2ID, info, false, info.ValidatorAppId)

	//TODO: it is not required
	if err != nil {
		return
	}

	return

}

//TODO: is this logic good?
func (s *Pool) RefreshWithInfo(info PoolInfo) {
	s.UpdateFromInfo(info)
}

func (s *Pool) Refresh() (err error) {

	info, err := GetPoolInfo(s.Client.Indexer, s.ValidatorAppID, s.Asset1.Id, s.Asset2.Id)
	//TODO:return error maybe
	if err != nil {
		return
	}

	s.UpdateFromInfo(info)

	return

}

//TODO: maybe None value could be -1 or 0
func (s *Pool) UpdateFromInfo(info PoolInfo) {

	//TODO: this is wrong
	if info.LiquidityAssetID != 0 {
		s.Exists = true
	}

	//TODO: asset Id to ID maybe
	s.LiquidityAsset = types.Asset{Id: info.LiquidityAssetID, Name: info.LiquidityAssetName, UnitName: "TMPOOL11", Decimals: 6}
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

func (s *Pool) GetLogicsig() (poolLogicsig algoTypes.LogicSig, err error) {

	poolLogicsig, err = contracts.GetPoolLogicsig(s.ValidatorAppID, s.Asset1.Id, s.Asset2.Id)

	//TODO: it is not required
	if err != nil {
		return
	}

	return

}

func (s *Pool) Address() (poolAddress algoTypes.Address, err error) {

	logicsig, err := s.GetLogicsig()

	if err != nil {
		return
	}

	poolAddress = crypto.AddressFromProgram(logicsig.Logic)

	return

}

//TODO: should return result be float64
func (s *Pool) Asset1Price() float64 {

	return float64(s.Asset2Reserves) / float64(s.Asset1Reserves)
}

func (s *Pool) Asset2Price() float64 {
	return float64(s.Asset1Reserves) / float64(s.Asset2Reserves)
}

func (s *Pool) Info() (poolInfo PoolInfo, err error) {

	address, err := s.Address()

	if err != nil {
		return
	}

	poolInfo = PoolInfo{
		Address:                         address,
		Asset1ID:                        s.Asset1.Id,
		Asset2ID:                        s.Asset2.Id,
		Asset1UnitName:                  s.Asset1.UnitName,
		Asset2UnitName:                  s.Asset2.UnitName,
		LiquidityAssetID:                s.LiquidityAsset.Id,
		LiquidityAssetName:              s.LiquidityAsset.Name,
		Asset1Reserves:                  s.Asset1Reserves,
		Asset2Reserves:                  s.Asset2Reserves,
		IssuedLiquidity:                 s.IssuedLiquidity,
		UnclaimedProtocolFees:           s.UnclaimedProtocolFees,
		OutstandingAsset1Amount:         s.OutstandingAsset1Amount,
		OutstandingAsset2Amount:         s.OutstandingAsset2Amount,
		OutstandingLiquidityAssetAmount: s.OutstandingLiquidityAssetAmount,
		LastRefreshedRound:              s.LastRefreshedRound,
	}

	return

}

func (s *Pool) Convert(amount types.AssetAmount) (assetAmount types.AssetAmount) {

	//TODO: maybe shorter way
	if amount.Asset == s.Asset1 {
		//TODO:maybe convert to int
		return types.AssetAmount{Asset: s.Asset2, Amount: uint64(float64(amount.Amount) * s.Asset1Price())}
	} else if amount.Asset == s.Asset2 {
		return types.AssetAmount{Asset: s.Asset1, Amount: uint64(float64(amount.Amount) * s.Asset2Price())}
	}

	return
}

//TODO: think about optional parameters
//TODO: default slippage
//TODO: check amountA and amountB if not nil so what
func (s *Pool) FetchMintQuote(amountA types.AssetAmount, amountB interface{}, slippage float64) (quote MintQuote, err error) {

	var amount1, amount2 interface{}
	var liquidityAssetAmount uint64

	if amountA.Asset == s.Asset1 {
		amount1 = amountA
	} else {
		amount1 = amountB
	}

	if amountA.Asset == s.Asset2 {
		amount2 = amountA
	} else {
		amount2 = amountB
	}

	err = s.Refresh()
	if err != nil {
		return
	}

	if !s.Exists {
		err = fmt.Errorf("pool has not been bootstrapped yet")
		return
	}

	//TODO: s.IssuedLiquidity could be None. think about a way
	//TODO: what about these type convertions
	if s.IssuedLiquidity > 0 {

		if amount1 == nil {
			amount1 = s.Convert(amount2.(types.AssetAmount))
		}

		if amount2 == nil {
			amount2 = s.Convert(amount1.(types.AssetAmount))
		}

		amount1, _ := amount1.(types.AssetAmount)
		amount2, _ := amount2.(types.AssetAmount)

		a := amount1.Amount * s.IssuedLiquidity / s.Asset1Reserves
		b := amount2.Amount * s.IssuedLiquidity / s.Asset2Reserves

		if a < b {
			liquidityAssetAmount = a
		} else {
			liquidityAssetAmount = b
		}

	} else {

		if amount1 == nil || amount2 == nil {
			err = fmt.Errorf("amounts required for both assets for first mint")
			return
		}

		amount1, _ := amount1.(types.AssetAmount)
		amount2, _ := amount2.(types.AssetAmount)

		liquidityAssetAmount = uint64(math.Sqrt(float64(amount1.Amount*amount2.Amount)) - 1000)
		slippage = 0

	}

	//TODO: maybe pointer
	quote = MintQuote{
		AmountsIn: map[types.Asset]types.AssetAmount{
			s.Asset1: amount1.(types.AssetAmount),
			s.Asset2: amount2.(types.AssetAmount),
		},
		LiquidityAssetAmount: types.AssetAmount{Asset: s.LiquidityAsset, Amount: liquidityAssetAmount},
		Slippage:             slippage,
	}

	return

}

//TODO: default value for slippage
func (s *Pool) FetchBurnQuote(liquidityAssetIn interface{}, slippage float64) (quote BurnQuote, err error) {

	var LiquidityAssetIn types.AssetAmount
	switch v := liquidityAssetIn.(type) {

	//TODO: maybe AssetAmount.Amount type is int
	case uint64:
		LiquidityAssetIn = types.AssetAmount{Asset: s.LiquidityAsset, Amount: v}
	case types.AssetAmount:
		liquidityAssetIn = v
	default:
		err = fmt.Errorf("unsupported type for liquidityAssetIn")
		return

	}

	err = s.Refresh()
	if err != nil {
		return
	}

	asset1Amount := (LiquidityAssetIn.Amount * s.Asset1Reserves) / s.IssuedLiquidity
	asset2Amount := (LiquidityAssetIn.Amount * s.Asset2Reserves) / s.IssuedLiquidity

	//TODO: maybe pointer
	quote = BurnQuote{
		AmountsOut: map[types.Asset]*types.AssetAmount{
			s.Asset1: {Asset: s.Asset1, Amount: asset1Amount},
			s.Asset2: {Asset: s.Asset2, Amount: asset2Amount},
		},
	}

	return

}

//TODO: we always return nil as error so do we need to return it
func (s *Pool) FetchFixedInputSwapQuote(amountIn types.AssetAmount, slippage float64) (quote SwapQuote, err error) {

	var assetOut types.Asset
	var inputSupply, outputSupply uint64

	assetIn := amountIn.Asset
	assetInAmount := amountIn.Amount

	err = s.Refresh()
	if err != nil {
		return
	}

	if assetIn == s.Asset1 {
		assetOut = s.Asset2
		inputSupply = s.Asset1Reserves
		outputSupply = s.Asset2Reserves
	} else {
		assetOut = s.Asset1
		inputSupply = s.Asset2Reserves
		outputSupply = s.Asset1Reserves
	}

	//TODO: how to implement this?
	// if !inputSupply || !outputSupply {
	// 	return SwapQuote{}, fmt.Errorf("pool has no liquidity")
	// }

	k := inputSupply * outputSupply
	assetInAmountMinusFee := (assetInAmount * 997) / 1000
	swapFees := assetInAmount - assetInAmountMinusFee
	assetOutAmount := outputSupply - (k / (inputSupply + uint64(assetInAmountMinusFee)))

	amountOut := types.AssetAmount{Asset: assetOut, Amount: assetOutAmount}

	//TODO: swap_fees is int but is set to AssetAmount in python code
	quote = SwapQuote{
		SwapType:  "fixed-input",
		AmountIn:  amountIn,
		AmountOut: amountOut,
		SwapFees:  types.AssetAmount{Asset: amountIn.Asset, Amount: swapFees},
		Slippage:  slippage,
	}

	return

}

//TODO: we always return nil as error so do we need to return it
func (s *Pool) FetchFixedOutputSwapQuote(amountOut types.AssetAmount, slippage float64) (quote SwapQuote, err error) {

	var assetIn types.Asset
	var inputSupply, outputSupply uint64

	assetOut := amountOut.Asset
	assetOutAmount := amountOut.Amount

	err = s.Refresh()
	if err != nil {
		return
	}

	if assetOut == s.Asset1 {
		assetIn = s.Asset2
		inputSupply = s.Asset2Reserves
		outputSupply = s.Asset1Reserves
	} else {
		assetIn = s.Asset1
		inputSupply = s.Asset1Reserves
		outputSupply = s.Asset2Reserves
	}

	k := inputSupply * outputSupply

	calculatedAmountInWithoutFee := (k / (outputSupply - uint64(assetOutAmount))) - inputSupply
	assetInAmount := calculatedAmountInWithoutFee * 1000 / 997
	swapFees := assetInAmount - calculatedAmountInWithoutFee

	amountIn := types.AssetAmount{Asset: assetIn, Amount: assetInAmount}

	//TODO: swap_fees is int but is set to AssetAmount in python code
	quote = SwapQuote{
		SwapType:  "fixed-output",
		AmountIn:  amountIn,
		AmountOut: amountOut,
		SwapFees:  types.AssetAmount{Asset: amountIn.Asset, Amount: swapFees},
		Slippage:  slippage,
	}

	return

}

//TODO: use address way of empty on others
func (s *Pool) PrepareSwapTransactions(amountIn types.AssetAmount, amountOut types.AssetAmount, swapType string, swapperAddress algoTypes.Address) (txnGroup utils.TransactionGroup, err error) {

	if swapperAddress.IsZero() {
		swapperAddress = s.Client.UserAddress
	}

	suggestedParams, err := s.Client.Algod.SuggestedParams().Do(context.Background())
	if err != nil {
		return
	}

	txnGroup, err = swap.PrepareSwapTransactions(
		s.ValidatorAppID,
		s.Asset1.Id,
		s.Asset2.Id,
		s.LiquidityAsset.Id,
		amountIn.Asset.Id,
		uint64(amountIn.Amount),
		uint64(amountOut.Amount),
		swapType,
		swapperAddress,
		suggestedParams,
	)

	//TODO: it is not required
	if err != nil {
		return
	}

	return

}

func (s *Pool) PrepareSwapTransactionsFromQuote(quote SwapQuote, swapperAddress algoTypes.Address) (txnGroup utils.TransactionGroup, err error) {
	amountIn, err := quote.AmountInWithSlippage()

	if err != nil {
		return
	}

	amountOut, err := quote.AmountOutWithSlippage()

	if err != nil {
		return
	}

	return s.PrepareSwapTransactions(amountIn, amountOut, quote.SwapType, swapperAddress)

}

func (s *Pool) PrepareBootstrapTransactions(poolerAddress algoTypes.Address) (txnGroup utils.TransactionGroup, err error) {

	if poolerAddress.IsZero() {
		poolerAddress = s.Client.UserAddress
	}

	suggestedParams, err := s.Client.Algod.SuggestedParams().Do(context.Background())
	if err != nil {
		return
	}

	txnGroup, err = bootstrap.PrepareBootstrapTransactions(s.ValidatorAppID,
		s.Asset1.Id,
		s.Asset2.Id,
		s.Asset1.UnitName,
		s.Asset2.UnitName,
		poolerAddress,
		suggestedParams)

	//TODO: it is not required
	if err != nil {
		return
	}

	return

}

//TODO: type dic[Asset] is dict[Asset,AssetAmount] in python code
func (s *Pool) PrepareMintTransactions(amountsIn map[types.Asset]types.AssetAmount, liquidityAssetAmount types.AssetAmount, poolerAddress algoTypes.Address) (txnGroup utils.TransactionGroup, err error) {

	if poolerAddress.IsZero() {
		poolerAddress = s.Client.UserAddress
	}

	asset1Amount := amountsIn[s.Asset1]
	asset2Amount := amountsIn[s.Asset2]

	suggestedParams, err := s.Client.Algod.SuggestedParams().Do(context.Background())
	if err != nil {
		return
	}

	txnGroup, err = mint.PrepareMintTransactions(s.ValidatorAppID,
		s.Asset1.Id,
		s.Asset2.Id,
		s.LiquidityAsset.Id,
		uint64(asset1Amount.Amount),
		uint64(asset2Amount.Amount),
		uint64(liquidityAssetAmount.Amount),
		poolerAddress,
		suggestedParams,
	)

	//TODO: it is not required
	if err != nil {
		return
	}

	return

}

func (s *Pool) PrepareMintTransactionsFromQuote(quote MintQuote, poolerAddress algoTypes.Address) (txnGroup utils.TransactionGroup, err error) {

	liquidityAssetAmount, err := quote.LiquidityAssetAmountWithSlippage()
	if err != nil {
		return
	}

	return s.PrepareMintTransactions(quote.AmountsIn, liquidityAssetAmount, poolerAddress)
}

func (s *Pool) PrepareBurnTransactions(liquidityAssetAmount interface{}, amountsOut map[types.Asset]types.AssetAmount, poolerAddress algoTypes.Address) (txnGroup utils.TransactionGroup, err error) {

	var LiquidityAssetAmount types.AssetAmount

	switch v := liquidityAssetAmount.(type) {
	case uint64:
		LiquidityAssetAmount = types.AssetAmount{Asset: s.LiquidityAsset, Amount: v}
	case types.AssetAmount:
		LiquidityAssetAmount = v
	default:
		return utils.TransactionGroup{}, fmt.Errorf("unsupported type for liquidityAssetAmount")
	}

	if poolerAddress.IsZero() {
		poolerAddress = s.Client.UserAddress
	}

	asset1Amount := amountsOut[s.Asset1]
	asset2Amount := amountsOut[s.Asset2]

	suggestedParams, err := s.Client.Algod.SuggestedParams().Do(context.Background())
	if err != nil {
		return
	}

	txnGroup, err = burn.PrepareBurnTransactions(
		s.ValidatorAppID,
		s.Asset1.Id,
		s.Asset2.Id,
		s.LiquidityAsset.Id,
		uint64(asset1Amount.Amount),
		uint64(asset2Amount.Amount),
		uint64(LiquidityAssetAmount.Amount),
		poolerAddress,
		suggestedParams,
	)

	//TODO: it is not required
	if err != nil {
		return
	}

	return

}

func (s *Pool) PrepareBurnTransactionsFromQuote(quote BurnQuote, poolerAddress algoTypes.Address) (txnGroup utils.TransactionGroup, err error) {

	amountsOut, err := quote.AmountsOutWithSlippage()

	if err != nil {
		return
	}

	return s.PrepareBurnTransactions(
		quote.LiquidityAssetAmount,
		amountsOut,
		poolerAddress,
	)

}

func (s *Pool) PrepareRedeemTransactions(amountOut types.AssetAmount, userAddress algoTypes.Address) (txnGroup utils.TransactionGroup, err error) {

	if userAddress.IsZero() {
		userAddress = s.Client.UserAddress
	}

	suggestedParams, err := s.Client.Algod.SuggestedParams().Do(context.Background())
	if err != nil {
		return
	}

	txnGroup, err = redeem.PrepareRedeemTransactions(
		s.ValidatorAppID,
		s.Asset1.Id,
		s.Asset2.Id,
		s.LiquidityAsset.Id,
		amountOut.Asset.Id,
		uint64(amountOut.Amount),
		userAddress,
		suggestedParams,
	)

	//TODO: it is not required
	if err != nil {
		return
	}

	return

}

func (s *Pool) PrepareLiquidityAssetOptinTransactions(userAddress algoTypes.Address) (txnGroup utils.TransactionGroup, err error) {

	if userAddress.IsZero() {
		userAddress = s.Client.UserAddress
	}

	suggestedParams, err := s.Client.Algod.SuggestedParams().Do(context.Background())
	if err != nil {
		return
	}

	txnGroup, err = optin.PrepareAssetOptinTransactions(
		s.LiquidityAsset.Id,
		userAddress,
		suggestedParams,
	)

	//TODO: it is not required
	if err != nil {
		return
	}

	return

}

func (s *Pool) PrepareRedeemFeesTransactions(amount uint64, creator algoTypes.Address, userAddress algoTypes.Address) (txnGroup utils.TransactionGroup, err error) {

	if userAddress.IsZero() {
		userAddress = s.Client.UserAddress
	}

	suggestedParams, err := s.Client.Algod.SuggestedParams().Do(context.Background())
	if err != nil {
		return
	}

	txnGroup, err = fees.PrepareRedeemFeesTransactions(
		s.ValidatorAppID,
		s.Asset1.Id,
		s.Asset2.Id,
		s.LiquidityAsset.Id,
		amount,
		creator,
		userAddress,
		suggestedParams,
	)

	//TODO: it is not required
	if err != nil {
		return
	}

	return
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

func (s *Pool) FetchExcessAmounts(userAddress algoTypes.Address) (excessAmounts map[types.Asset]types.AssetAmount, err error) {
	if userAddress.IsZero() {
		userAddress = s.Client.UserAddress
	}

	address, err := s.Address()
	if err != nil {
		return nil, err
	}

	fetchedExcessAmounts, err := s.Client.FetchExcessAmounts(userAddress)
	if err != nil {
		return nil, err
	}

	if val, ok := fetchedExcessAmounts[address.String()]; ok {
		return val, nil
	} else {
		return map[types.Asset]types.AssetAmount{}, nil
	}

}

func (s *Pool) FetchPoolPosition(poolerAddress algoTypes.Address) (map[interface{}]interface{}, error) {

	if poolerAddress.IsZero() {
		poolerAddress = s.Client.UserAddress
	}

	_, accountInfo, err := s.Client.Indexer.LookupAccountByID(poolerAddress.String()).Do(context.Background())
	if err != nil {
		return nil, err
	}

	Assets := make(map[uint64]models.AssetHolding)
	for _, a := range accountInfo.Assets {
		Assets[a.AssetId] = a
	}

	var liquidityAssetAmount uint64
	if val, ok := Assets[s.LiquidityAsset.Id]; ok {
		liquidityAssetAmount = val.Amount
	} else {
		liquidityAssetAmount = 0
	}

	quote, err := s.FetchBurnQuote(liquidityAssetAmount, 0.05)
	if err != nil {
		return nil, err
	}

	//TODO: return type
	//TODO: pointer or not
	return map[interface{}]interface{}{
		s.Asset1: *quote.AmountsOut[s.Asset1],
		s.Asset2: *quote.AmountsOut[s.Asset2],
		"share":  liquidityAssetAmount / s.IssuedLiquidity,
	}, nil

}

//TODO: return types is different
func (s *Pool) FetchState(key interface{}) (state interface{}, err error) {

	address, err := s.Address()
	if err != nil {
		return
	}

	accountInfo, err := s.Client.Algod.AccountInformation(address.String()).Do(context.Background())

	if err != nil {
		return
	}

	if len(accountInfo.AppsLocalState) == 0 {
		err = fmt.Errorf("accountInfo.AppsLocalState len is 0")
		return
	}

	// validatorAppID := accountInfo.AppsLocalState[0].Id

	validatorAppState := make(map[string]models.TealValue)

	for _, x := range accountInfo.AppsLocalState[0].KeyValue {
		validatorAppState[x.Key] = x.Value
	}

	if key != nil {
		return utils.GetStateInt(validatorAppState, key), nil
	}

	return validatorAppState, nil

}