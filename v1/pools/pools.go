package pools

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"strconv"

	"github.com/soheil555/tinyman-mobile-sdk/types"
	"github.com/soheil555/tinyman-mobile-sdk/utils"
	"github.com/soheil555/tinyman-mobile-sdk/v1/bootstrap"
	"github.com/soheil555/tinyman-mobile-sdk/v1/burn"
	"github.com/soheil555/tinyman-mobile-sdk/v1/client"
	"github.com/soheil555/tinyman-mobile-sdk/v1/contracts"
	"github.com/soheil555/tinyman-mobile-sdk/v1/fees"
	"github.com/soheil555/tinyman-mobile-sdk/v1/mint"
	"github.com/soheil555/tinyman-mobile-sdk/v1/optin"
	"github.com/soheil555/tinyman-mobile-sdk/v1/redeem"
	"github.com/soheil555/tinyman-mobile-sdk/v1/swap"

	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
	"github.com/algorand/go-algorand-sdk/crypto"
	algoTypes "github.com/algorand/go-algorand-sdk/types"
)

type PoolInfo struct {
	Address                         string `json:"address"`
	Asset1Id                        int    `json:"asset1-id"`
	Asset2Id                        int    `json:"asset2-id"`
	Asset1UnitName                  string `json:"asset1-unit-name"`
	Asset2UnitName                  string `json:"asset2-unit-name"`
	LiquidityAssetId                int    `json:"liquidity-asset-id"`
	LiquidityAssetName              string `json:"liquidity-asset-name"`
	Asset1Reserves                  string `json:"asset1-reserves"`
	Asset2Reserves                  string `json:"asset2-reserves"`
	IssuedLiquidity                 string `json:"issued-liquidity"`
	UnclaimedProtocolFees           string `json:"unclaimed-protocol-fees"`
	OutstandingAsset1Amount         string `json:"outstanding-asset1-amount"`
	OutstandingAsset2Amount         string `json:"outstanding-asset2-amount"`
	OutstandingLiquidityAssetAmount string `json:"outstanding-liquidity-asset-amount"`
	ValidatorAppId                  int    `json:"validator-app-id"`
	AlgoBalance                     string `json:"algo-balance"`
	Round                           int    `json:"round"`
	LastRefreshedRound              int    `json:"last-refreshed-round"`
}

func GetPoolInfo(client *client.TinymanClient, validatorAppID, asset1ID, asset2ID int) (poolInfo *PoolInfo, err error) {

	poolLogicsig, err := contracts.GetPoolLogicsig(validatorAppID, asset1ID, asset2ID)
	if err != nil {
		return
	}

	poolAddress := crypto.AddressFromProgram(poolLogicsig.Logic)

	_, accountInfo, err := client.LookupAccountByID(poolAddress.String())
	if err != nil {
		return
	}

	return GetPoolInfoFromAccountInfo(accountInfo)

}

// not compatible with go-mobile
func GetPoolInfoFromAccountInfo(accountInfo models.Account) (poolInfo *PoolInfo, err error) {

	if len(accountInfo.AppsLocalState) == 0 {
		return
	}

	validatorAppID := int(accountInfo.AppsLocalState[0].Id)

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
	liquidityAssetID := int(liquidityAsset.Index)

	key1 := []byte("o")
	key1 = append(key1, utils.IntToBytes(int(asset1Id))...)
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

	asset1ReservesBig := big.NewInt(int64(asset1Reserves))
	asset2ReservesBig := big.NewInt(int64(asset2Reserves))
	issuedLiquidityBig := big.NewInt(int64(issuedLiquidity))
	accountAmountBig := big.NewInt(int64(accountInfo.Amount))

	unclaimedProtocolFeesBig := big.NewInt(int64(unclaimedProtocolFees))
	outstandingAsset1AmountBig := big.NewInt(int64(outstandingAsset1Amount))
	outstandingAsset2AmountBig := big.NewInt(int64(outstandingAsset2Amount))
	outstandingLiquidityAssetAmountBig := big.NewInt(int64(outstandingLiquidityAssetAmount))

	poolInfo = &PoolInfo{
		Address:                         poolAddress.String(),
		Asset1Id:                        asset1Id,
		Asset2Id:                        asset2Id,
		LiquidityAssetId:                liquidityAssetID,
		LiquidityAssetName:              liquidityAsset.Params.Name,
		Asset1Reserves:                  asset1ReservesBig.String(),
		Asset2Reserves:                  asset2ReservesBig.String(),
		IssuedLiquidity:                 issuedLiquidityBig.String(),
		UnclaimedProtocolFees:           unclaimedProtocolFeesBig.String(),
		OutstandingAsset1Amount:         outstandingAsset1AmountBig.String(),
		OutstandingAsset2Amount:         outstandingAsset2AmountBig.String(),
		OutstandingLiquidityAssetAmount: outstandingLiquidityAssetAmountBig.String(),
		ValidatorAppId:                  validatorAppID,
		AlgoBalance:                     accountAmountBig.String(),
		Round:                           int(accountInfo.Round),
	}

	return

}

func GetExcessAssetKey(poolAddress string, assetID int) (key []byte, err error) {
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
	SwapType  string             `json:"swap-type"`
	AmountIn  *types.AssetAmount `json:"amount-in"`
	AmountOut *types.AssetAmount `json:"amount-out"`
	SwapFees  *types.AssetAmount `json:"swap-fees"`
	Slippage  float64            `json:"slippage"`
}

func (s *SwapQuote) AmountOutWithSlippage() (assetAmount *types.AssetAmount, err error) {

	if s.SwapType == "fixed-output" {
		return s.AmountOut, nil
	}

	assetAmount, err = s.AmountOut.Sub(s.AmountOut.Mul(s.Slippage))

	return

}

func (s *SwapQuote) AmountInWithSlippage() (assetAmount *types.AssetAmount, err error) {

	if s.SwapType == "fixed-input" {
		return s.AmountIn, nil
	}

	assetAmount, err = s.AmountIn.Add(s.AmountIn.Mul(s.Slippage))

	return

}

func (s *SwapQuote) Price() float64 {

	sAmountIn := utils.NewBigFloatString(s.AmountIn.Amount)
	sAmountOut := utils.NewBigFloatString(s.AmountOut.Amount)

	price := new(big.Float)
	price.Quo(sAmountOut, sAmountIn)

	priceFloat64, _ := price.Float64()

	return priceFloat64

}

func (s *SwapQuote) PriceWithSlippage() (priceWithSlippage float64, err error) {

	amountInWithSlippage, err := s.AmountInWithSlippage()

	if err != nil {
		return
	}

	amountOutWithSlippage, err := s.AmountOutWithSlippage()

	if err != nil {
		return
	}

	sAmountInWithSlippage := utils.NewBigFloatString(amountInWithSlippage.Amount)
	sAmountOutWithSlippage := utils.NewBigFloatString(amountOutWithSlippage.Amount)

	priceWithSlippageBig := new(big.Float)
	priceWithSlippageBig.Quo(sAmountOutWithSlippage, sAmountInWithSlippage)

	priceWithSlippage, _ = priceWithSlippageBig.Float64()

	return

}

//TODO: in python code AmountsIn is dict[AssetAmount]
type MintQuote struct {
	amountsIn            map[int]string     // map[asset.id][assetAmount.Amount]
	LiquidityAssetAmount *types.AssetAmount `json:"liquidity-asset-amount"`
	Slippage             float64            `json:"slippage"`
}

func (s *MintQuote) GetAmountsInStr() (string, error) {

	amountsIn, err := json.Marshal(s.amountsIn)
	return string(amountsIn), err

}

func (s *MintQuote) GetAmountsIn() map[int]string {

	return s.amountsIn

}

//TODO: in python code it return int
func (s *MintQuote) LiquidityAssetAmountWithSlippage() (assetAmount *types.AssetAmount, err error) {
	assetAmount, err = s.LiquidityAssetAmount.Sub(s.LiquidityAssetAmount.Mul(s.Slippage))
	return
}

type BurnQuote struct {
	amountsOut           map[int]string     // map[asset.id][assetAmount.Amount]
	LiquidityAssetAmount *types.AssetAmount `json:"liquidity-asset-amount"`
	Slippage             float64            `json:"slippage"`
}

func (s *BurnQuote) GetAmountsOutStr() (amountsOutStr string, err error) {

	amountsOutBytes, err := json.Marshal(s.amountsOut)
	if err != nil {
		return
	}

	amountsOutStr = string(amountsOutBytes)
	return

}

func (s *BurnQuote) GetAmountsOut() map[int]string {

	return s.amountsOut

}

func (s *BurnQuote) AmountsOutWithSlippage() (amountsOutWithSlippage map[int]string, err error) {

	amountsOutWithSlippage = make(map[int]string)

	for k := range s.amountsOut {

		amountsOut := utils.NewBigFloatString(s.amountsOut[k])
		slippage := big.NewFloat(s.Slippage)

		helper := new(big.Float)
		helper.Mul(amountsOut, slippage)

		amountOutWithSlippageInt, _ := new(big.Float).Sub(amountsOut, helper).Int(nil)
		amountsOutWithSlippage[k] = amountOutWithSlippageInt.String()

	}

	return
}

func (s *BurnQuote) AmountsOutWithSlippageStr() (amountsOutWithSlippageStr string, err error) {

	amountsOutWithSlippage, err := s.AmountsOutWithSlippage()
	if err != nil {
		return
	}

	amountsOutWithSlippageBytes, err := json.Marshal(amountsOutWithSlippage)
	if err != nil {
		return
	}

	amountsOutWithSlippageStr = string(amountsOutWithSlippageBytes)
	return

}

type Pool struct {
	Client                          *client.TinymanClient `json:"client"`
	ValidatorAppId                  int                   `json:"validator-app-id"`
	Asset1                          *types.Asset          `json:"asset1"`
	Asset2                          *types.Asset          `json:"asset2"`
	Exists                          bool                  `json:"exists"`
	LiquidityAsset                  *types.Asset          `json:"liquidity-asset"`
	Asset1Reserves                  string                `json:"asset1-reserves"`
	Asset2Reserves                  string                `json:"asset2-reserves"`
	IssuedLiquidity                 string                `json:"issued-liquidity"`
	UnclaimedProtocolFees           string                `json:"unclaimed-protocol-fees"`
	OutstandingAsset1Amount         string                `json:"outstanding-asset1-amount"`
	OutstandingAsset2Amount         string                `json:"outstanding-asset2-amount"`
	OutstandingLiquidityAssetAmount string                `json:"outstanding-liquidity-asset-amount"`
	LastRefreshedRound              int                   `json:"last-refreshed-round"`
	AlgoBalance                     string                `json:"algo-balance"`
	MinBalance                      int                   `json:"min-balance"`
}

//TODO: is validatorID == 0 a valid ID
func NewPool(client *client.TinymanClient, assetA, assetB *types.Asset, info *PoolInfo, fetch bool, validatorAppId int) (pool *Pool, err error) {

	pool = new(Pool)

	if assetA == nil || assetB == nil {
		err = fmt.Errorf("assetA and assetB are required")
		return
	}

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

	} else if info != nil {

		pool.UpdateFromInfo(info)

	}

	return

}

// not compatible with go-mobile
func NewPoolFromAccountInfo(accountInfo models.Account, client *client.TinymanClient) (pool *Pool, err error) {

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

	info, err := GetPoolInfo(s.Client, s.ValidatorAppId, s.Asset1.Id, s.Asset2.Id)

	if err != nil || reflect.ValueOf(info).IsZero() {
		return
	}

	s.UpdateFromInfo(info)

	return

}

func (s *Pool) UpdateFromInfo(info *PoolInfo) {

	//TODO: LiquidityAssetID is an ASA(Algorand Standard Asset). 0 is not a valid ASA ID
	if info.LiquidityAssetId != 0 {
		s.Exists = true
	}

	s.LiquidityAsset = &types.Asset{Id: info.LiquidityAssetId, Name: info.LiquidityAssetName, UnitName: "TMPOOL11", Decimals: 6}
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

		algoBalance := utils.NewBigIntString(s.AlgoBalance)
		outstandingAsset2Amount := utils.NewBigIntString(s.OutstandingAsset2Amount)
		minBalance := big.NewInt(int64(s.MinBalance))

		asset2Reserves := new(big.Int)
		asset2Reserves.Sub(algoBalance, minBalance)
		asset2Reserves.Sub(asset2Reserves, outstandingAsset2Amount)

		s.Asset2Reserves = asset2Reserves.String()

	}

}

func (s *Pool) GetLogicsig() (poolLogicsig *types.LogicSig, err error) {

	poolLogicsig, err = contracts.GetPoolLogicsig(s.ValidatorAppId, s.Asset1.Id, s.Asset2.Id)

	return

}

func (s *Pool) Address() (poolAddress string, err error) {

	logicsig, err := s.GetLogicsig()

	if err != nil {
		return
	}

	poolAddress = crypto.AddressFromProgram(logicsig.Logic).String()

	return

}

func (s *Pool) Asset1Price() float64 {

	asset2Reserves := utils.NewBigFloatString(s.Asset2Reserves)
	asset1Reserves := utils.NewBigFloatString(s.Asset1Reserves)

	asset1Price := new(big.Float)
	asset1Price.Quo(asset2Reserves, asset1Reserves)

	asset1PriceFloat64, _ := asset1Price.Float64()
	return asset1PriceFloat64

}

func (s *Pool) Asset2Price() float64 {

	asset2Reserves := utils.NewBigFloatString(s.Asset2Reserves)
	asset1Reserves := utils.NewBigFloatString(s.Asset1Reserves)

	asset1Price := new(big.Float)
	asset1Price.Quo(asset1Reserves, asset2Reserves)

	asset1PriceFloat64, _ := asset1Price.Float64()

	return asset1PriceFloat64

}

func (s *Pool) Info() (poolInfo *PoolInfo, err error) {

	address, err := s.Address()

	if err != nil {
		return
	}

	poolInfo = &PoolInfo{
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

func (s *Pool) Convert(amount *types.AssetAmount) (assetAmount *types.AssetAmount) {

	helper := utils.NewBigFloatString(amount.Amount)

	if *amount.Asset == *s.Asset1 {

		asset1Price := big.NewFloat(s.Asset1Price())
		Amount, _ := new(big.Float).Mul(helper, asset1Price).Int(nil)

		assetAmount = &types.AssetAmount{Asset: s.Asset2, Amount: Amount.String()}

	} else if *amount.Asset == *s.Asset2 {

		asset2Price := big.NewFloat(s.Asset2Price())
		Amount, _ := new(big.Float).Mul(helper, asset2Price).Int(nil)

		assetAmount = &types.AssetAmount{Asset: s.Asset1, Amount: Amount.String()}

	}

	return
}

func (s *Pool) FetchMintQuote(amountA, amountB *types.AssetAmount, slippage float64) (quote *MintQuote, err error) {

	var amount1, amount2 *types.AssetAmount
	var liquidityAssetAmount string

	if *amountA.Asset == *s.Asset1 {
		amount1 = amountA
	} else {
		amount1 = amountB
	}

	if *amountA.Asset == *s.Asset2 {
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

	issuedLiquidity := utils.NewBigFloatString(s.IssuedLiquidity)
	if issuedLiquidity.Sign() > 0 {

		if amount1 == nil {
			amount1 = s.Convert(amount2)
		}

		if amount2 == nil {
			amount2 = s.Convert(amount1)
		}

		amount1Amount := utils.NewBigFloatString(amount1.Amount)
		amount2Amount := utils.NewBigFloatString(amount2.Amount)
		asset1Reserves := utils.NewBigFloatString(s.Asset1Reserves)
		asset2Reserves := utils.NewBigFloatString(s.Asset2Reserves)

		helper1 := new(big.Float).Mul(amount1Amount, issuedLiquidity)
		helper2 := new(big.Float).Mul(amount2Amount, issuedLiquidity)

		a := new(big.Float).Quo(helper1, asset1Reserves)
		b := new(big.Float).Quo(helper2, asset2Reserves)

		if a.Cmp(b) < 0 {

			a, _ := a.Int(nil)
			liquidityAssetAmount = a.String()

		} else {

			b, _ := b.Int(nil)
			liquidityAssetAmount = b.String()

		}

	} else {

		if amount1 == nil || amount2 == nil {
			err = fmt.Errorf("amounts required for both assets for first mint")
			return
		}

		amount1Amount := utils.NewBigFloatString(amount1.Amount)
		amount2Amount := utils.NewBigFloatString(amount2.Amount)

		helper := new(big.Float)
		helper.Mul(amount1Amount, amount2Amount)
		helper.Sqrt(helper)
		helper.Sub(helper, big.NewFloat(1000))

		helperInt, _ := helper.Int(nil)
		liquidityAssetAmount = helperInt.String()

		slippage = 0

	}

	quote = &MintQuote{
		amountsIn: map[int]string{
			s.Asset1.Id: amount1.Amount,
			s.Asset2.Id: amount2.Amount,
		},
		LiquidityAssetAmount: &types.AssetAmount{Asset: s.LiquidityAsset, Amount: liquidityAssetAmount},
		Slippage:             slippage,
	}

	return

}

func (s *Pool) FetchMintQuoteWithDefaultSlippage(amountA, amountB *types.AssetAmount) (quote *MintQuote, err error) {

	return s.FetchMintQuote(amountA, amountB, 0.05)

}

func (s *Pool) FetchBurnQuote(liquidityAssetIn *types.AssetAmount, slippage float64) (quote *BurnQuote, err error) {

	err = s.Refresh()
	if err != nil {
		return
	}

	liquidityAssetInAmount := utils.NewBigFloatString(liquidityAssetIn.Amount)
	asset1Reserves := utils.NewBigFloatString(s.Asset1Reserves)
	asset2Reserves := utils.NewBigFloatString(s.Asset2Reserves)
	issuedLiquidity := utils.NewBigFloatString(s.IssuedLiquidity)

	helper1 := new(big.Float).Mul(liquidityAssetInAmount, asset1Reserves)
	helper2 := new(big.Float).Mul(liquidityAssetInAmount, asset2Reserves)

	asset1Amount := new(big.Float).Quo(helper1, issuedLiquidity)
	asset2Amount := new(big.Float).Quo(helper2, issuedLiquidity)

	asset1AmountInt, _ := asset1Amount.Int(nil)
	asset2AmountInt, _ := asset2Amount.Int(nil)

	quote = &BurnQuote{
		amountsOut: map[int]string{
			s.Asset1.Id: asset1AmountInt.String(),
			s.Asset2.Id: asset2AmountInt.String(),
		},
		LiquidityAssetAmount: liquidityAssetIn,
		Slippage:             slippage,
	}

	return

}

func (s *Pool) FetchBurnQuoteWithDefaultSlippage(liquidityAssetIn *types.AssetAmount) (quote *BurnQuote, err error) {

	return s.FetchBurnQuote(liquidityAssetIn, 0.05)

}

func (s *Pool) FetchFixedInputSwapQuote(amountIn *types.AssetAmount, slippage float64) (quote *SwapQuote, err error) {

	var assetOut *types.Asset
	var inputSupply, outputSupply string

	assetIn := amountIn.Asset
	assetInAmount := amountIn.Amount

	err = s.Refresh()
	if err != nil {
		return
	}

	if *assetIn == *s.Asset1 {
		assetOut = s.Asset2
		inputSupply = s.Asset1Reserves
		outputSupply = s.Asset2Reserves
	} else {
		assetOut = s.Asset1
		inputSupply = s.Asset2Reserves
		outputSupply = s.Asset1Reserves
	}

	inputSupplyBig := utils.NewBigFloatString(inputSupply)
	outputSupplyBig := utils.NewBigFloatString(outputSupply)

	if inputSupplyBig.Sign() == 0 || outputSupplyBig.Sign() == 0 {
		err = fmt.Errorf("pool has no liquidity")
		return
	}

	k := new(big.Float).Mul(inputSupplyBig, outputSupplyBig)
	assetInAmountBig := utils.NewBigFloatString(assetInAmount)

	helper1 := new(big.Float).Mul(assetInAmountBig, big.NewFloat(997))
	assetInAmountMinusFee := new(big.Float).Quo(helper1, big.NewFloat(1000))

	swapFees := new(big.Float).Sub(assetInAmountBig, assetInAmountMinusFee)

	helper2 := new(big.Float)
	helper2.Add(inputSupplyBig, assetInAmountMinusFee)
	helper2.Quo(k, helper2)

	assetOutAmount := new(big.Float).Sub(outputSupplyBig, helper2)
	assetOutAmountInt, _ := assetOutAmount.Int(nil)

	amountOut := types.AssetAmount{Asset: assetOut, Amount: assetOutAmountInt.String()}

	quote = &SwapQuote{
		SwapType:  "fixed-input",
		AmountIn:  amountIn,
		AmountOut: &amountOut,
		SwapFees:  &types.AssetAmount{Asset: amountIn.Asset, Amount: swapFees.String()},
		Slippage:  slippage,
	}

	return

}

func (s *Pool) FetchFixedInputSwapQuoteWithDefaultSlippage(amountIn *types.AssetAmount) (quote *SwapQuote, err error) {
	return s.FetchFixedInputSwapQuote(amountIn, 0.05)
}

func (s *Pool) FetchFixedOutputSwapQuote(amountOut *types.AssetAmount, slippage float64) (quote *SwapQuote, err error) {

	var assetIn *types.Asset
	var inputSupply, outputSupply string

	assetOut := amountOut.Asset
	assetOutAmount := amountOut.Amount

	err = s.Refresh()
	if err != nil {
		return
	}

	if *assetOut == *s.Asset1 {
		assetIn = s.Asset2
		inputSupply = s.Asset2Reserves
		outputSupply = s.Asset1Reserves
	} else {
		assetIn = s.Asset1
		inputSupply = s.Asset1Reserves
		outputSupply = s.Asset2Reserves
	}

	inputSupplyBig := utils.NewBigFloatString(inputSupply)
	outputSupplyBig := utils.NewBigFloatString(outputSupply)
	assetOutAmountBig := utils.NewBigFloatString(assetOutAmount)

	k := new(big.Float).Mul(inputSupplyBig, outputSupplyBig)

	helper := new(big.Float).Sub(outputSupplyBig, assetOutAmountBig)
	helper.Quo(k, helper)

	calculatedAmountInWithoutFee := new(big.Float).Sub(helper, inputSupplyBig)

	assetInAmount := new(big.Float).Mul(calculatedAmountInWithoutFee, big.NewFloat(1000))
	assetInAmount.Quo(assetInAmount, big.NewFloat(997))

	swapFees := new(big.Float).Sub(assetInAmount, calculatedAmountInWithoutFee)

	swapFeesInt, _ := swapFees.Int(nil)
	assetInAmountInt, _ := assetInAmount.Int(nil)

	amountIn := types.AssetAmount{Asset: assetIn, Amount: assetInAmountInt.String()}

	quote = &SwapQuote{
		SwapType:  "fixed-output",
		AmountIn:  &amountIn,
		AmountOut: amountOut,
		SwapFees:  &types.AssetAmount{Asset: amountIn.Asset, Amount: swapFeesInt.String()},
		Slippage:  slippage,
	}

	return

}

func (s *Pool) FetchFixedOutputSwapQuoteWithDefaultSlippage(amountOut *types.AssetAmount) (quote *SwapQuote, err error) {
	return s.FetchFixedOutputSwapQuote(amountOut, 0.05)
}

func (s *Pool) PrepareSwapTransactions(amountIn, amountOut *types.AssetAmount, swapType string, swapperAddress string) (txnGroup *utils.TransactionGroup, err error) {

	if len(swapperAddress) == 0 {
		swapperAddress = s.Client.UserAddress
	}

	swapper, err := algoTypes.DecodeAddress(swapperAddress)
	if err != nil {
		return
	}

	algoSuggestedParams, err := s.Client.SuggestedParams()
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

	txnGroup, err = swap.PrepareSwapTransactions(
		s.ValidatorAppId,
		s.Asset1.Id,
		s.Asset2.Id,
		s.LiquidityAsset.Id,
		amountIn.Asset.Id,
		amountIn.Amount,
		amountOut.Amount,
		swapType,
		swapper.String(),
		suggestedParams,
	)

	return

}

func (s *Pool) PrepareSwapTransactionsFromQuote(quote *SwapQuote, swapperAddress string) (txnGroup *utils.TransactionGroup, err error) {

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

func (s *Pool) PrepareBootstrapTransactions(poolerAddress string) (txnGroup *utils.TransactionGroup, err error) {

	if len(poolerAddress) == 0 {
		poolerAddress = s.Client.UserAddress
	}

	pooler, err := algoTypes.DecodeAddress(poolerAddress)
	if err != nil {
		return
	}

	algoSuggestedParams, err := s.Client.SuggestedParams()
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

	txnGroup, err = bootstrap.PrepareBootstrapTransactions(s.ValidatorAppId,
		s.Asset1.Id,
		s.Asset2.Id,
		s.Asset1.UnitName,
		s.Asset2.UnitName,
		pooler.String(),
		suggestedParams)

	return

}

//TODO: type dic[Asset] is dict[Asset,AssetAmount] in python code
func (s *Pool) PrepareMintTransactions(amountsInStr string, liquidityAssetAmount *types.AssetAmount, poolerAddress string) (txnGroup *utils.TransactionGroup, err error) {

	amountsIn := make(map[int]string)
	err = json.Unmarshal([]byte(amountsInStr), &amountsIn)
	if err != nil {
		return
	}

	if len(poolerAddress) == 0 {
		poolerAddress = s.Client.UserAddress
	}

	pooler, err := algoTypes.DecodeAddress(poolerAddress)
	if err != nil {
		return
	}

	asset1Amount := amountsIn[s.Asset1.Id]
	asset2Amount := amountsIn[s.Asset2.Id]

	algoSuggestedParams, err := s.Client.SuggestedParams()
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

	txnGroup, err = mint.PrepareMintTransactions(s.ValidatorAppId,
		s.Asset1.Id,
		s.Asset2.Id,
		s.LiquidityAsset.Id,
		asset1Amount,
		asset2Amount,
		liquidityAssetAmount.Amount,
		pooler.String(),
		suggestedParams,
	)

	return

}

func (s *Pool) PrepareMintTransactionsFromQuote(quote *MintQuote, poolerAddress string) (txnGroup *utils.TransactionGroup, err error) {

	liquidityAssetAmount, err := quote.LiquidityAssetAmountWithSlippage()
	if err != nil {
		return
	}

	amountsIn, err := quote.GetAmountsInStr()
	if err != nil {
		return
	}

	return s.PrepareMintTransactions(amountsIn, liquidityAssetAmount, poolerAddress)
}

func (s *Pool) PrepareBurnTransactions(liquidityAssetAmount *types.AssetAmount, amountsOut map[int]string, poolerAddress string) (txnGroup *utils.TransactionGroup, err error) {

	if len(poolerAddress) == 0 {
		poolerAddress = s.Client.UserAddress
	}

	pooler, err := algoTypes.DecodeAddress(poolerAddress)
	if err != nil {
		return
	}

	asset1Amount := amountsOut[s.Asset1.Id]
	asset2Amount := amountsOut[s.Asset2.Id]

	algoSuggestedParams, err := s.Client.SuggestedParams()
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

	txnGroup, err = burn.PrepareBurnTransactions(
		s.ValidatorAppId,
		s.Asset1.Id,
		s.Asset2.Id,
		s.LiquidityAsset.Id,
		asset1Amount,
		asset2Amount,
		liquidityAssetAmount.Amount,
		pooler.String(),
		suggestedParams,
	)

	return

}

func (s *Pool) PrepareBurnTransactionsWithAmountsOutStr(liquidityAssetAmount *types.AssetAmount, amountsOutStr, poolerAddress string) (txnGroup *utils.TransactionGroup, err error) {

	amountsOut := make(map[int]string)
	err = json.Unmarshal([]byte(amountsOutStr), &amountsOut)
	if err != nil {
		return
	}

	return s.PrepareBurnTransactions(liquidityAssetAmount, amountsOut, poolerAddress)

}

func (s *Pool) PrepareBurnTransactionsFromQuote(quote *BurnQuote, poolerAddress string) (txnGroup *utils.TransactionGroup, err error) {

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

func (s *Pool) PrepareRedeemTransactions(amountOut *types.AssetAmount, userAddress string) (txnGroup *utils.TransactionGroup, err error) {

	if len(userAddress) == 0 {
		userAddress = s.Client.UserAddress
	}

	user, err := algoTypes.DecodeAddress(userAddress)
	if err != nil {
		return
	}

	algoSuggestedParams, err := s.Client.SuggestedParams()
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

	txnGroup, err = redeem.PrepareRedeemTransactions(
		s.ValidatorAppId,
		s.Asset1.Id,
		s.Asset2.Id,
		s.LiquidityAsset.Id,
		amountOut.Asset.Id,
		amountOut.Amount,
		user.String(),
		suggestedParams,
	)

	return

}

func (s *Pool) PrepareLiquidityAssetOptinTransactions(userAddress string) (txnGroup *utils.TransactionGroup, err error) {

	if len(userAddress) == 0 {
		userAddress = s.Client.UserAddress
	}

	user, err := algoTypes.DecodeAddress(userAddress)
	if err != nil {
		return
	}

	algoSuggestedParams, err := s.Client.SuggestedParams()
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

	txnGroup, err = optin.PrepareAssetOptinTransactions(
		s.LiquidityAsset.Id,
		user.String(),
		suggestedParams,
	)

	return

}

func (s *Pool) PrepareRedeemFeesTransactions(amount, creatorAddress, userAddress string) (txnGroup *utils.TransactionGroup, err error) {

	if len(userAddress) == 0 {
		userAddress = s.Client.UserAddress
	}

	user, err := algoTypes.DecodeAddress(userAddress)
	if err != nil {
		return
	}

	creator, err := algoTypes.DecodeAddress(userAddress)
	if err != nil {
		return
	}

	algoSuggestedParams, err := s.Client.SuggestedParams()
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

	txnGroup, err = fees.PrepareRedeemFeesTransactions(
		s.ValidatorAppId,
		s.Asset1.Id,
		s.Asset2.Id,
		s.LiquidityAsset.Id,
		amount,
		creator.String(),
		user.String(),
		suggestedParams,
	)

	return
}

func (s *Pool) GetMinimumBalance() int {

	const (
		MIN_BALANCE_PER_ACCOUNT       int = 100000
		MIN_BALANCE_PER_ASSET         int = 100000
		MIN_BALANCE_PER_APP           int = 100000
		MIN_BALANCE_PER_APP_BYTESLICE int = 50000
		MIN_BALANCE_PER_APP_UINT      int = 28500
	)

	var numAssets int
	if s.Asset2.Id == 0 {
		numAssets = 2
	} else {
		numAssets = 3
	}

	var numCreatedApps int = 0
	var numLocalApps int = 1
	var totalUnits int = 16
	var totalByteslices int = 0

	total := MIN_BALANCE_PER_ACCOUNT + (MIN_BALANCE_PER_ASSET * numAssets) + (MIN_BALANCE_PER_APP * (numCreatedApps + numLocalApps)) + MIN_BALANCE_PER_APP_UINT*totalUnits + MIN_BALANCE_PER_APP_BYTESLICE*totalByteslices
	return total
}

func (s *Pool) FetchExcessAmounts(userAddress string) (excessAmounts map[int]string, err error) {

	if len(userAddress) == 0 {
		userAddress = s.Client.UserAddress
	}

	user, err := algoTypes.DecodeAddress(userAddress)
	if err != nil {
		return
	}

	address, err := s.Address()
	if err != nil {
		return
	}

	fetchedExcessAmounts := make(map[string]map[int]string)
	fetchedExcessAmountsStr, err := s.Client.FetchExcessAmounts(user.String())
	if err != nil {
		return
	}

	json.Unmarshal([]byte(fetchedExcessAmountsStr), &fetchedExcessAmounts)

	if val, ok := fetchedExcessAmounts[address]; ok {
		return val, nil
	} else {
		return
	}

}

func (s *Pool) FetchExcessAmountsStr(userAddress string) (excessAmountsStr string, err error) {

	excessAmounts, err := s.FetchExcessAmounts(userAddress)
	if err != nil {
		return
	}

	var excessAmountsBytes []byte
	excessAmountsBytes, err = json.Marshal(excessAmounts)
	if err != nil {
		return
	}
	excessAmountsStr = string(excessAmountsBytes)
	return

}

func (s *Pool) FetchPoolPosition(poolerAddress string) (poolPosition map[string]string, err error) {

	if len(poolerAddress) == 0 {
		poolerAddress = s.Client.UserAddress
	}

	pooler, err := algoTypes.DecodeAddress(poolerAddress)
	if err != nil {
		return
	}

	_, accountInfo, err := s.Client.LookupAccountByID(pooler.String())
	if err != nil {
		return
	}

	Assets := make(map[uint64]models.AssetHolding)
	for _, a := range accountInfo.Assets {
		Assets[a.AssetId] = a
	}

	var liquidityAssetAmount string
	if val, ok := Assets[uint64(s.LiquidityAsset.Id)]; ok {
		liquidityAssetAmount = big.NewInt(int64(val.Amount)).String()
	} else {
		liquidityAssetAmount = "0"
	}

	liquidityAssetIn := &types.AssetAmount{Asset: s.LiquidityAsset, Amount: liquidityAssetAmount}

	quote, err := s.FetchBurnQuoteWithDefaultSlippage(liquidityAssetIn)
	if err != nil {
		return
	}

	liquidityAssetAmountBig := utils.NewBigFloatString(liquidityAssetAmount)
	issuedLiquidityBig := utils.NewBigFloatString(s.IssuedLiquidity)

	share := new(big.Float).Quo(liquidityAssetAmountBig, issuedLiquidityBig)

	amountsOut := quote.GetAmountsOut()
	poolPosition = map[string]string{
		strconv.Itoa(s.Asset1.Id):         amountsOut[s.Asset1.Id],
		strconv.Itoa(s.Asset2.Id):         amountsOut[s.Asset2.Id],
		strconv.Itoa(s.LiquidityAsset.Id): quote.LiquidityAssetAmount.Amount,
		"share":                           share.String(),
	}

	return

}

func (s *Pool) FetchPoolPositionStr(poolerAddress string) (poolPositionStr string, err error) {

	poolPosition, err := s.FetchPoolPosition(poolerAddress)
	if err != nil {
		return
	}

	poolPositionBytes, err := json.Marshal(poolPosition)
	if err != nil {
		return
	}

	poolPositionStr = string(poolPositionBytes)
	return

}

func (s *Pool) FetchState() (validatorAppState map[string]models.TealValue, err error) {

	address, err := s.Address()
	if err != nil {
		return
	}

	accountInfo, err := s.Client.AccountInformation(address)

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

	return

}

func (s *Pool) FetchStateStr() (validatorAppStateStr string, err error) {

	validatorAppState, err := s.FetchState()
	if err != nil {
		return
	}

	validatorAppStateBytes, err := json.Marshal(validatorAppState)
	if err != nil {
		return
	}

	validatorAppStateStr = string(validatorAppStateBytes)
	return

}

func (s *Pool) FetchStateWithKey(key interface{}) (state int, err error) {

	address, err := s.Address()
	if err != nil {
		return
	}

	accountInfo, err := s.Client.AccountInformation(address)

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

func (s *Pool) FetchStateWithStrKey(key string) (state int, err error) {
	return s.FetchStateWithKey(key)
}

func (s *Pool) FetchStateWithBytesKey(key []byte) (state int, err error) {
	return s.FetchStateWithKey(key)
}
