package pools

import (
	"context"
	b64 "encoding/base64"
	"fmt"
	"math"
	"math/big"
	"reflect"
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

//TODO: round vs lastRefreshedRound
type PoolInfo struct {
	Address                         algoTypes.Address
	Asset1Id                        uint64
	Asset2Id                        uint64
	Asset1UnitName                  string
	Asset2UnitName                  string
	LiquidityAssetId                uint64
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

func GetPoolInfo(indexer *indexer.Client, validatorAppID, asset1ID, asset2ID int) (poolInfo *PoolInfo, err error) {

	poolLogicsig, err := contracts.GetPoolLogicsig(validatorAppID, asset1ID, asset2ID)
	if err != nil {
		return
	}

	poolAddress := crypto.AddressFromProgram(poolLogicsig.Logic)

	_, accountInfo, err := indexer.LookupAccountByID(poolAddress.String()).Do(context.Background())
	if err != nil {
		return
	}

	return GetPoolInfoFromAccountInfo(accountInfo)

}

func GetPoolInfoFromAccountInfo(accountInfo models.Account) (poolInfo *PoolInfo, err error) {

	if len(accountInfo.AppsLocalState) == 0 {
		return
	}

	validatorAppID := accountInfo.AppsLocalState[0].Id

	validatorAppState := make(map[string]models.TealValue)

	for _, x := range accountInfo.AppsLocalState[0].KeyValue {
		validatorAppState[x.Key] = x.Value
	}

	asset1Id := utils.GetStateInt(validatorAppState, "a1")
	asset2Id := utils.GetStateInt(validatorAppState, "a2")

	poolLogicsig, err := contracts.GetPoolLogicsig(validatorAppID, asset1Id, asset2Id)

	if err != nil {
		return
	}

	poolAddress := crypto.AddressFromProgram(poolLogicsig.Logic)

	if accountInfo.Address != poolAddress.String() {
		err = fmt.Errorf("accountInfo address is not equal to poolAddress")
		return
	}

	asset1Reserves := utils.GetStateInt(validatorAppState, "s1")
	asset2Reserves := utils.GetStateInt(validatorAppState, "s2")
	issuedLiquidity := utils.GetStateInt(validatorAppState, "ilt")
	unclaimedProtocolFees := utils.GetStateInt(validatorAppState, "p")

	liquidityAsset := accountInfo.CreatedAssets[0]
	liquidityAssetID := liquidityAsset.Index

	key1 := []byte("o")
	key1 = append(key1, utils.IntToBytes(asset1Id)...)
	encodedKey1 := make([]byte, b64.StdEncoding.EncodedLen(len(key1)))

	b64.StdEncoding.Encode(encodedKey1, key1)

	outstandingAsset1Amount := utils.GetStateInt(validatorAppState, encodedKey1)

	key2 := []byte("o")
	key2 = append(key2, utils.IntToBytes(asset2Id)...)
	encodedKey2 := make([]byte, b64.StdEncoding.EncodedLen(len(key2)))

	outstandingAsset2Amount := utils.GetStateInt(validatorAppState, encodedKey2)

	key3 := []byte("o")
	key3 = append(key3, utils.IntToBytes(liquidityAssetID)...)
	encodedKey3 := make([]byte, b64.StdEncoding.EncodedLen(len(key3)))

	outstandingLiquidityAssetAmount := utils.GetStateInt(validatorAppState, encodedKey3)

	poolInfo = PoolInfo{
		Address:                         poolAddress,
		Asset1Id:                        asset1Id,
		Asset2Id:                        asset2Id,
		LiquidityAssetId:                liquidityAsset.Index,
		LiquidityAssetName:              liquidityAsset.Params.Name,
		Asset1Reserves:                  asset1Reserves,
		Asset2Reserves:                  asset2Reserves,
		IssuedLiquidity:                 issuedLiquidity,
		UnclaimedProtocolFees:           unclaimedProtocolFees,
		OutstandingAsset1Amount:         outstandingAsset1Amount,
		OutstandingAsset2Amount:         outstandingAsset2Amount,
		OutstandingLiquidityAssetAmount: outstandingLiquidityAssetAmount,
		ValidatorAppId:                  validatorAppID,
		AlgoBalance:                     accountInfo.Amount,
		Round:                           accountInfo.Round,
	}

	return

}

func GetExcessAssetKey(poolAddress string, assetID uint64) (key []byte, err error) {
	a, err := algoTypes.DecodeAddress(poolAddress)
	if err != nil {
		return
	}

	key = append(key, a[:]...)
	key = append(key, byte('e'))
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

	return

}

func (s *SwapQuote) AmountInWithSlippage() (assetAmount types.AssetAmount, err error) {

	if s.SwapType == "fixed-input" {
		return s.AmountIn, nil
	}

	assetAmount, err = s.AmountIn.Add(s.AmountIn.Mul(s.Slippage))

	return

}

func (s *SwapQuote) Price() float64 {
	return float64(s.AmountOut.Amount) / float64(s.AmountIn.Amount)
}

func (s *SwapQuote) PriceWithSlippage() (priceWithSlippage float64, err error) {

	amountOutWithSlippage, err := s.AmountOutWithSlippage()

	if err != nil {
		return
	}

	amountInWithSlippage, err := s.AmountInWithSlippage()

	if err != nil {
		return
	}

	priceWithSlippage = float64(amountOutWithSlippage.Amount) / float64(amountInWithSlippage.Amount)

	return

}

//TODO: in python code AmountsIn is dict[AssetAmount]
type MintQuote struct {
	AmountsIn            map[types.Asset]*types.AssetAmount
	LiquidityAssetAmount types.AssetAmount
	Slippage             float64
}

//TODO: in python code it return int
func (s *MintQuote) LiquidityAssetAmountWithSlippage() (assetAmount types.AssetAmount, err error) {
	assetAmount, err = s.LiquidityAssetAmount.Sub(s.LiquidityAssetAmount.Mul(s.Slippage))

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

type Pool struct {
	Client                          client.TinymanClient
	ValidatorAppId                  uint64
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

//TODO: is validatorID == 0 a valid ID
//TODO: assetA and assetB could be either int or Asset. but what if input is some type like uint64
func NewPool(client client.TinymanClient, assetA types.Asset, assetB types.Asset, info PoolInfo, fetch bool, validatorAppId uint64) (pool Pool, err error) {

	pool.Client = client

	if validatorAppId == 0 {
		pool.ValidatorAppId = client.ValidatorAppId
	} else {
		pool.ValidatorAppId = validatorAppId
	}

	if assetA.Id > assetB.Id {

		pool.Asset1 = assetA
		pool.Asset2 = assetB

	} else {

		pool.Asset1 = assetB
		pool.Asset2 = assetA

	}

	if fetch {

		err = pool.Refresh()
		if err != nil {
			return
		}

	} else if !reflect.ValueOf(info).IsZero() {

		pool.UpdateFromInfo(info)

	}

	return

}

func NewPoolFromAccountInfo(accountInfo models.Account, client client.TinymanClient) (pool Pool, err error) {

	info, err := GetPoolInfoFromAccountInfo(accountInfo)

	if err != nil {
		return
	}

	asset1, err := client.FetchAsset(info.Asset1Id)

	if err != nil {
		return
	}

	asset2, err := client.FetchAsset(info.Asset2Id)
	if err != nil {
		return
	}

	pool, err = NewPool(client, asset1, asset2, info, true, info.ValidatorAppId)

	return

}

func (s *Pool) Refresh() (err error) {

	info, err := GetPoolInfo(s.Client.Indexer, s.ValidatorAppId, s.Asset1.Id, s.Asset2.Id)

	if err != nil || reflect.ValueOf(info).IsZero() {
		return
	}

	s.UpdateFromInfo(info)

	return

}

func (s *Pool) UpdateFromInfo(info PoolInfo) {

	//TODO: LiquidityAssetID is an ASA(Algorand Standard Asset). 0 is not a valid ASA ID
	if info.LiquidityAssetId != 0 {
		s.Exists = true
	}

	s.LiquidityAsset = types.Asset{Id: info.LiquidityAssetId, Name: info.LiquidityAssetName, UnitName: "TMPOOL11", Decimals: 6}
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

	poolLogicsig, err = contracts.GetPoolLogicsig(s.ValidatorAppId, s.Asset1.Id, s.Asset2.Id)

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
		Asset1Id:                        s.Asset1.Id,
		Asset2Id:                        s.Asset2.Id,
		Asset1UnitName:                  s.Asset1.UnitName,
		Asset2UnitName:                  s.Asset2.UnitName,
		LiquidityAssetId:                s.LiquidityAsset.Id,
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

	if amount.Asset == s.Asset1 {
		assetAmount = types.AssetAmount{Asset: s.Asset2, Amount: uint64(float64(amount.Amount) * s.Asset1Price())}
	} else if amount.Asset == s.Asset2 {
		assetAmount = types.AssetAmount{Asset: s.Asset1, Amount: uint64(float64(amount.Amount) * s.Asset2Price())}
	}

	return
}

func (s *Pool) FetchMintQuote(amountA *types.AssetAmount, amountB *types.AssetAmount, slippage float64) (quote MintQuote, err error) {

	var amount1, amount2 types.AssetAmount
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

	//TODO: in python code 0 is invalid so I think this is correct
	if s.IssuedLiquidity > 0 {

		if reflect.ValueOf(amount1).IsZero() {
			amount1 = s.Convert(amount2)
		}

		if reflect.ValueOf(amount2).IsZero() {
			amount2 = s.Convert(amount1)
		}

		a := amount1.Amount * s.IssuedLiquidity / s.Asset1Reserves
		b := amount2.Amount * s.IssuedLiquidity / s.Asset2Reserves

		if a < b {
			liquidityAssetAmount = a
		} else {
			liquidityAssetAmount = b
		}

	} else {

		if reflect.ValueOf(amount1).IsZero() || reflect.ValueOf(amount2).IsZero() {
			err = fmt.Errorf("amounts required for both assets for first mint")
			return
		}

		liquidityAssetAmount = uint64(math.Sqrt(float64(amount1.Amount*amount2.Amount)) - 1000)
		slippage = 0

	}

	quote = MintQuote{
		AmountsIn: map[types.Asset]*types.AssetAmount{
			s.Asset1: &amount1,
			s.Asset2: &amount2,
		},
		LiquidityAssetAmount: types.AssetAmount{Asset: s.LiquidityAsset, Amount: liquidityAssetAmount},
		Slippage:             slippage,
	}

	return

}

func (s *Pool) FetchMintQuoteWithDefaultSlippage(amountA types.AssetAmount, amountB types.AssetAmount) (quote MintQuote, err error) {

	return s.FetchMintQuote(amountA, amountB, 0.05)

}

//TODO: should I handle int for liquidityAssetIn
func (s *Pool) FetchBurnQuote(liquidityAssetIn types.AssetAmount, slippage float64) (quote BurnQuote, err error) {

	err = s.Refresh()
	if err != nil {
		return
	}

	asset1Amount := (liquidityAssetIn.Amount * s.Asset1Reserves) / s.IssuedLiquidity
	asset2Amount := (liquidityAssetIn.Amount * s.Asset2Reserves) / s.IssuedLiquidity

	quote = BurnQuote{
		AmountsOut: map[types.Asset]*types.AssetAmount{
			s.Asset1: {Asset: s.Asset1, Amount: asset1Amount},
			s.Asset2: {Asset: s.Asset2, Amount: asset2Amount},
		},
		LiquidityAssetAmount: liquidityAssetIn,
		Slippage:             slippage,
	}

	return

}

func (s *Pool) FetchBurnQuoteWithDefaultSlippage(liquidityAssetIn types.AssetAmount) (quote BurnQuote, err error) {

	return s.FetchBurnQuote(liquidityAssetIn, 0.05)

}

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

	if inputSupply == 0 || outputSupply == 0 {
		err = fmt.Errorf("pool has no liquidity")
		return
	}

	k := new(big.Int).Mul(big.NewInt(int64(inputSupply)), big.NewInt(int64(outputSupply)))
	assetInAmountMinusFee := assetInAmount * 997 / 1000
	swapFees := assetInAmount - assetInAmountMinusFee

	tmp := new(big.Int).Div(k, big.NewInt(int64(inputSupply+assetInAmountMinusFee)))
	assetOutAmount := new(big.Int).Sub(big.NewInt(int64(outputSupply)), tmp)

	amountOut := types.AssetAmount{Asset: assetOut, Amount: assetOutAmount.Uint64()}

	quote = SwapQuote{
		SwapType:  "fixed-input",
		AmountIn:  amountIn,
		AmountOut: amountOut,
		SwapFees:  types.AssetAmount{Asset: amountIn.Asset, Amount: swapFees},
		Slippage:  slippage,
	}

	return

}

func (s *Pool) FetchFixedInputSwapQuoteWithDefaultSlippage(amountIn types.AssetAmount) (quote SwapQuote, err error) {
	return s.FetchFixedInputSwapQuote(amountIn, 0.05)
}

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

	k := new(big.Int).Mul(big.NewInt(int64(inputSupply)), big.NewInt(int64(outputSupply)))

	tmp := new(big.Int).Div(k, big.NewInt(int64(outputSupply-assetOutAmount)))
	calculatedAmountInWithoutFee := new(big.Int).Sub(tmp, big.NewInt(int64(inputSupply)))

	assetInAmount := new(big.Int).Mul(calculatedAmountInWithoutFee, big.NewInt(1000))
	assetInAmount.Div(assetInAmount, big.NewInt(997))

	swapFees := new(big.Int).Sub(assetInAmount, calculatedAmountInWithoutFee)

	amountIn := types.AssetAmount{Asset: assetIn, Amount: assetInAmount.Uint64()}

	quote = SwapQuote{
		SwapType:  "fixed-output",
		AmountIn:  amountIn,
		AmountOut: amountOut,
		SwapFees:  types.AssetAmount{Asset: amountIn.Asset, Amount: swapFees.Uint64()},
		Slippage:  slippage,
	}

	return

}

func (s *Pool) FetchFixedOutputSwapQuoteWithDefaultSlippage(amountOut types.AssetAmount) (quote SwapQuote, err error) {
	return s.FetchFixedOutputSwapQuote(amountOut, 0.05)
}

func (s *Pool) PrepareSwapTransactions(amountIn types.AssetAmount, amountOut types.AssetAmount, swapType string, swapperAddress algoTypes.Address) (txnGroup utils.TransactionGroup, err error) {

	if swapperAddress.IsZero() {
		swapperAddress = s.Client.UserAddress
	}

	suggestedParams, err := s.Client.Algod.SuggestedParams().Do(context.Background())
	if err != nil {
		return
	}

	txnGroup, err = swap.PrepareSwapTransactions(
		s.ValidatorAppId,
		s.Asset1.Id,
		s.Asset2.Id,
		s.LiquidityAsset.Id,
		amountIn.Asset.Id,
		amountIn.Amount,
		amountOut.Amount,
		swapType,
		swapperAddress,
		suggestedParams,
	)

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

	txnGroup, err = bootstrap.PrepareBootstrapTransactions(s.ValidatorAppId,
		s.Asset1.Id,
		s.Asset2.Id,
		s.Asset1.UnitName,
		s.Asset2.UnitName,
		poolerAddress,
		suggestedParams)

	return

}

//TODO: type dic[Asset] is dict[Asset,AssetAmount] in python code
func (s *Pool) PrepareMintTransactions(amountsIn map[types.Asset]*types.AssetAmount, liquidityAssetAmount types.AssetAmount, poolerAddress algoTypes.Address) (txnGroup utils.TransactionGroup, err error) {

	if poolerAddress.IsZero() {
		poolerAddress = s.Client.UserAddress
	}

	asset1Amount := amountsIn[s.Asset1]
	asset2Amount := amountsIn[s.Asset2]

	suggestedParams, err := s.Client.Algod.SuggestedParams().Do(context.Background())
	if err != nil {
		return
	}

	txnGroup, err = mint.PrepareMintTransactions(s.ValidatorAppId,
		s.Asset1.Id,
		s.Asset2.Id,
		s.LiquidityAsset.Id,
		asset1Amount.Amount,
		asset2Amount.Amount,
		liquidityAssetAmount.Amount,
		poolerAddress,
		suggestedParams,
	)

	return

}

func (s *Pool) PrepareMintTransactionsFromQuote(quote MintQuote, poolerAddress algoTypes.Address) (txnGroup utils.TransactionGroup, err error) {

	liquidityAssetAmount, err := quote.LiquidityAssetAmountWithSlippage()
	if err != nil {
		return
	}

	return s.PrepareMintTransactions(quote.AmountsIn, liquidityAssetAmount, poolerAddress)
}

func (s *Pool) PrepareBurnTransactions(liquidityAssetAmount types.AssetAmount, amountsOut map[types.Asset]types.AssetAmount, poolerAddress algoTypes.Address) (txnGroup utils.TransactionGroup, err error) {

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
		s.ValidatorAppId,
		s.Asset1.Id,
		s.Asset2.Id,
		s.LiquidityAsset.Id,
		asset1Amount.Amount,
		asset2Amount.Amount,
		liquidityAssetAmount.Amount,
		poolerAddress,
		suggestedParams,
	)

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
		s.ValidatorAppId,
		s.Asset1.Id,
		s.Asset2.Id,
		s.LiquidityAsset.Id,
		amountOut.Asset.Id,
		amountOut.Amount,
		userAddress,
		suggestedParams,
	)

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
		s.ValidatorAppId,
		s.Asset1.Id,
		s.Asset2.Id,
		s.LiquidityAsset.Id,
		amount,
		creator,
		userAddress,
		suggestedParams,
	)

	return
}

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
		return
	}

	fetchedExcessAmounts, err := s.Client.FetchExcessAmounts(userAddress)
	if err != nil {
		return
	}

	if val, ok := fetchedExcessAmounts[address.String()]; ok {
		return val, nil
	} else {
		return
	}

}

func (s *Pool) FetchPoolPosition(poolerAddress algoTypes.Address) (poolPosition map[types.Asset]types.AssetAmount, share float64, err error) {

	if poolerAddress.IsZero() {
		poolerAddress = s.Client.UserAddress
	}

	_, accountInfo, err := s.Client.Indexer.LookupAccountByID(poolerAddress.String()).Do(context.Background())
	if err != nil {
		return
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

	liquidityAssetIn := types.AssetAmount{Asset: s.LiquidityAsset, Amount: liquidityAssetAmount}

	quote, err := s.FetchBurnQuoteWithDefaultSlippage(liquidityAssetIn)
	if err != nil {
		return
	}

	poolPosition = map[types.Asset]types.AssetAmount{
		s.Asset1:         *quote.AmountsOut[s.Asset1],
		s.Asset2:         *quote.AmountsOut[s.Asset2],
		s.LiquidityAsset: quote.LiquidityAssetAmount,
	}

	share = float64(liquidityAssetAmount) / float64(s.IssuedLiquidity)

	return

}

func (s *Pool) FetchState() (validatorAppState map[string]models.TealValue, err error) {

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
	validatorAppState = make(map[string]models.TealValue)

	for _, x := range accountInfo.AppsLocalState[0].KeyValue {
		validatorAppState[x.Key] = x.Value
	}

	return validatorAppState, nil

}

func (s *Pool) FetchStateWithKey(key interface{}) (state uint64, err error) {

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

	return utils.GetStateInt(validatorAppState, key), nil

}
