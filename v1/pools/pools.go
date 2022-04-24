package pools

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"strconv"

	"github.com/soheil555/tinyman-mobile-sdk/assets"
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
	Address                         string
	Asset1Id                        int
	Asset2Id                        int
	Asset1UnitName                  string
	Asset2UnitName                  string
	LiquidityAssetId                int
	LiquidityAssetName              string
	Asset1Reserves                  string
	Asset2Reserves                  string
	IssuedLiquidity                 string
	UnclaimedProtocolFees           string
	OutstandingAsset1Amount         string
	OutstandingAsset2Amount         string
	OutstandingLiquidityAssetAmount string
	ValidatorAppId                  int
	AlgoBalance                     string
	Round                           int
	LastRefreshedRound              int
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

	Asset1Reserves := big.NewInt(int64(asset1Reserves))
	Asset2Reserves := big.NewInt(int64(asset2Reserves))
	IssuedLiquidity := big.NewInt(int64(issuedLiquidity))
	AccountAmount := big.NewInt(int64(accountInfo.Amount))

	UnclaimedProtocolFees := big.NewInt(int64(unclaimedProtocolFees))
	OutstandingAsset1Amount := big.NewInt(int64(outstandingAsset1Amount))
	OutstandingAsset2Amount := big.NewInt(int64(outstandingAsset2Amount))
	OutstandingLiquidityAssetAmount := big.NewInt(int64(outstandingLiquidityAssetAmount))

	poolInfo = &PoolInfo{
		Address:                         poolAddress.String(),
		Asset1Id:                        asset1Id,
		Asset2Id:                        asset2Id,
		LiquidityAssetId:                liquidityAssetID,
		LiquidityAssetName:              liquidityAsset.Params.Name,
		Asset1Reserves:                  Asset1Reserves.String(),
		Asset2Reserves:                  Asset2Reserves.String(),
		IssuedLiquidity:                 IssuedLiquidity.String(),
		UnclaimedProtocolFees:           UnclaimedProtocolFees.String(),
		OutstandingAsset1Amount:         OutstandingAsset1Amount.String(),
		OutstandingAsset2Amount:         OutstandingAsset2Amount.String(),
		OutstandingLiquidityAssetAmount: OutstandingLiquidityAssetAmount.String(),
		ValidatorAppId:                  validatorAppID,
		AlgoBalance:                     AccountAmount.String(),
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
	SwapType  string
	AmountIn  *assets.AssetAmount
	AmountOut *assets.AssetAmount
	SwapFees  *assets.AssetAmount
	Slippage  float64
}

func (s *SwapQuote) AmountOutWithSlippage() (assetAmount *assets.AssetAmount, err error) {

	if s.SwapType == "fixed-output" {
		return s.AmountOut, nil
	}

	assetAmount, err = s.AmountOut.Sub(s.AmountOut.Mul(s.Slippage))

	return

}

func (s *SwapQuote) AmountInWithSlippage() (assetAmount *assets.AssetAmount, err error) {

	if s.SwapType == "fixed-input" {
		return s.AmountIn, nil
	}

	assetAmount, err = s.AmountIn.Add(s.AmountIn.Mul(s.Slippage))

	return

}

func (s *SwapQuote) Price() float64 {

	sAmountIn, ok := new(big.Float).SetString(s.AmountIn.Amount)
	if !ok {
		return 0
	}

	sAmountOut, ok := new(big.Float).SetString(s.AmountOut.Amount)
	if !ok {
		return 0
	}

	num := new(big.Float).Quo(sAmountOut, sAmountIn)
	numFloat, _ := num.Float64()

	return numFloat
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

	sAmountInWithSlippage, ok := new(big.Float).SetString(amountInWithSlippage.Amount)
	if !ok {

		return 0, fmt.Errorf("failed to convert amountIn to float")
	}

	sAmountOutWithSlippage, ok := new(big.Float).SetString(amountOutWithSlippage.Amount)
	if !ok {
		return 0, fmt.Errorf("failed to convert amountOut to float")
	}

	num := new(big.Float).Quo(sAmountOutWithSlippage, sAmountInWithSlippage)
	priceWithSlippage, _ = num.Float64()

	return

}

//TODO: in python code AmountsIn is dict[AssetAmount]
type MintQuote struct {
	amountsIn            map[int]string // map[asset.id][assetAmount.Amount]
	LiquidityAssetAmount *assets.AssetAmount
	Slippage             float64
}

func (s *MintQuote) GetAmountsInStr() (string, error) {

	amountsIn, err := json.Marshal(s.amountsIn)
	return string(amountsIn), err

}

func (s *MintQuote) GetAmountsIn() map[int]string {

	return s.amountsIn

}

//TODO: in python code it return int
func (s *MintQuote) LiquidityAssetAmountWithSlippage() (assetAmount *assets.AssetAmount, err error) {
	assetAmount, err = s.LiquidityAssetAmount.Sub(s.LiquidityAssetAmount.Mul(s.Slippage))
	return
}

type BurnQuote struct {
	amountsOut           map[int]string // map[asset.id][assetAmount.Amount]
	LiquidityAssetAmount *assets.AssetAmount
	Slippage             float64
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

func (s *BurnQuote) AmountsOutWithSlippage() (amountsOutWithSlippage string, err error) {

	out := make(map[int]string)

	for k := range s.amountsOut {

		amountsOut, _ := new(big.Float).SetString(s.amountsOut[k])
		slippage := big.NewFloat(s.Slippage)

		tmp := new(big.Float).Mul(amountsOut, slippage)

		amountOutWithSlippageUint, _ := new(big.Float).Sub(amountsOut, tmp).Int(nil)

		if err != nil {
			return
		}

		out[k] = amountOutWithSlippageUint.String()
	}

	amountsOutWithSlippageBytes, err := json.Marshal(out)
	if err != nil {
		return
	}

	amountsOutWithSlippage = string(amountsOutWithSlippageBytes)
	return
}

type Pool struct {
	Client                          *client.TinymanClient
	ValidatorAppId                  int
	Asset1                          *assets.Asset
	Asset2                          *assets.Asset
	Exists                          bool
	LiquidityAsset                  *assets.Asset
	Asset1Reserves                  string
	Asset2Reserves                  string
	IssuedLiquidity                 string
	UnclaimedProtocolFees           string
	OutstandingAsset1Amount         string
	OutstandingAsset2Amount         string
	OutstandingLiquidityAssetAmount string
	LastRefreshedRound              int
	AlgoBalance                     string
	MinBalance                      int
}

//TODO: is validatorID == 0 a valid ID
func NewPool(client *client.TinymanClient, assetA, assetB *assets.Asset, info *PoolInfo, fetch bool, validatorAppId int) (pool *Pool, err error) {

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

	s.LiquidityAsset = &assets.Asset{Id: info.LiquidityAssetId, Name: info.LiquidityAssetName, UnitName: "TMPOOL11", Decimals: 6}
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

		algoBalance, _ := new(big.Int).SetString(s.AlgoBalance, 10)
		OutstandingAsset2Amount, _ := new(big.Int).SetString(s.OutstandingAsset2Amount, 10)
		minBalance := big.NewInt(int64(s.MinBalance))

		tmp := new(big.Int).Sub(algoBalance, minBalance)
		tmp.Sub(tmp, OutstandingAsset2Amount)

		s.Asset2Reserves = tmp.String()

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

	asset2Reserves, _ := new(big.Float).SetString(s.Asset2Reserves)
	asset1Reserves, _ := new(big.Float).SetString(s.Asset1Reserves)

	num := new(big.Float).Quo(asset2Reserves, asset1Reserves)
	numFloat, _ := num.Float64()

	return numFloat
}

func (s *Pool) Asset2Price() float64 {

	asset2Reserves, _ := new(big.Float).SetString(s.Asset2Reserves)
	asset1Reserves, _ := new(big.Float).SetString(s.Asset1Reserves)

	num := new(big.Float).Quo(asset1Reserves, asset2Reserves)
	numFloat, _ := num.Float64()

	return numFloat
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

func (s *Pool) Convert(amount *assets.AssetAmount) (assetAmount *assets.AssetAmount) {

	tmp, _ := new(big.Float).SetString(amount.Amount)

	if *amount.Asset == *s.Asset1 {

		asset1Price := big.NewFloat(s.Asset1Price())
		Amount, _ := new(big.Float).Mul(tmp, asset1Price).Int(nil)

		assetAmount = &assets.AssetAmount{Asset: s.Asset2, Amount: Amount.String()}

	} else if *amount.Asset == *s.Asset2 {

		asset2Price := big.NewFloat(s.Asset2Price())
		Amount, _ := new(big.Float).Mul(tmp, asset2Price).Int(nil)

		assetAmount = &assets.AssetAmount{Asset: s.Asset1, Amount: Amount.String()}
	}

	return
}

func (s *Pool) FetchMintQuote(amountA, amountB *assets.AssetAmount, slippage float64) (quote *MintQuote, err error) {

	var amount1, amount2 *assets.AssetAmount
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

	issuedLiquidity, ok := new(big.Int).SetString(s.IssuedLiquidity, 10)
	if ok && issuedLiquidity.Cmp(big.NewInt(0)) > 0 {

		if amount1 == nil {
			amount1 = s.Convert(amount2)
		}

		if amount2 == nil {
			amount2 = s.Convert(amount1)
		}

		amount1Amount, ok := new(big.Int).SetString(amount1.Amount, 10)
		if !ok {
			err = fmt.Errorf("failed to convert amount1.Amount to int")
			return
		}

		amount2Amount, ok := new(big.Int).SetString(amount2.Amount, 10)
		if !ok {
			err = fmt.Errorf("failed to convert amount2.Amount to int")
			return
		}

		asset1Reserves, ok := new(big.Int).SetString(s.Asset1Reserves, 10)
		if !ok {
			err = fmt.Errorf("failed to convert Asset1Reserves to int")
			return
		}

		asset2Reserves, ok := new(big.Int).SetString(s.Asset2Reserves, 10)
		if !ok {
			err = fmt.Errorf("failed to convert Asset2Reserves to int")
			return
		}

		tmp1 := new(big.Int).Mul(amount1Amount, issuedLiquidity)
		tmp2 := new(big.Int).Mul(amount2Amount, issuedLiquidity)

		a := new(big.Float).Quo(new(big.Float).SetInt(tmp1), new(big.Float).SetInt(asset1Reserves))
		b := new(big.Float).Quo(new(big.Float).SetInt(tmp2), new(big.Float).SetInt(asset2Reserves))

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

		amount1Amount, _ := new(big.Float).SetString(amount1.Amount)
		amount2Amount, _ := new(big.Float).SetString(amount2.Amount)

		tmp := new(big.Float).Mul(amount1Amount, amount2Amount)
		tmp.Sqrt(tmp)
		tmp.Sub(tmp, big.NewFloat(1000))

		tmpInt, _ := tmp.Int(nil)

		liquidityAssetAmount = tmpInt.String()
		slippage = 0

	}

	quote = &MintQuote{
		amountsIn: map[int]string{
			s.Asset1.Id: amount1.Amount,
			s.Asset2.Id: amount2.Amount,
		},
		LiquidityAssetAmount: &assets.AssetAmount{Asset: s.LiquidityAsset, Amount: liquidityAssetAmount},
		Slippage:             slippage,
	}

	return

}

func (s *Pool) FetchMintQuoteWithDefaultSlippage(amountA, amountB *assets.AssetAmount) (quote *MintQuote, err error) {

	return s.FetchMintQuote(amountA, amountB, 0.05)

}

func (s *Pool) FetchBurnQuote(liquidityAssetIn *assets.AssetAmount, slippage float64) (quote *BurnQuote, err error) {

	err = s.Refresh()
	if err != nil {
		return
	}

	liquidityAssetInAmount, _ := new(big.Int).SetString(liquidityAssetIn.Amount, 10)
	asset1Reserves, _ := new(big.Int).SetString(s.Asset1Reserves, 10)
	asset2Reserves, _ := new(big.Int).SetString(s.Asset2Reserves, 10)
	issuedLiquidity, _ := new(big.Int).SetString(s.IssuedLiquidity, 10)

	tmp1 := new(big.Int).Mul(liquidityAssetInAmount, asset1Reserves)
	tmp2 := new(big.Int).Mul(liquidityAssetInAmount, asset2Reserves)

	asset1Amount := new(big.Float).Quo(new(big.Float).SetInt(tmp1), new(big.Float).SetInt(issuedLiquidity))
	asset2Amount := new(big.Float).Quo(new(big.Float).SetInt(tmp2), new(big.Float).SetInt(issuedLiquidity))

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

func (s *Pool) FetchBurnQuoteWithDefaultSlippage(liquidityAssetIn *assets.AssetAmount) (quote *BurnQuote, err error) {

	return s.FetchBurnQuote(liquidityAssetIn, 0.05)

}

func (s *Pool) FetchFixedInputSwapQuote(amountIn *assets.AssetAmount, slippage float64) (quote *SwapQuote, err error) {

	var assetOut *assets.Asset
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

	InputSupply, _ := new(big.Int).SetString(inputSupply, 10)
	OutputSupply, _ := new(big.Int).SetString(outputSupply, 10)

	if InputSupply.Cmp(big.NewInt(0)) == 0 || OutputSupply.Cmp(big.NewInt(0)) == 0 {
		err = fmt.Errorf("pool has no liquidity")
		return
	}

	k := new(big.Int).Mul(InputSupply, OutputSupply)
	AssetInAmount, _ := new(big.Int).SetString(assetInAmount, 10)

	tmp := new(big.Int).Mul(AssetInAmount, big.NewInt(997))

	assetInAmountMinusFee := new(big.Float).Quo(new(big.Float).SetInt(tmp), big.NewFloat(1000))
	swapFees := new(big.Float).Sub(new(big.Float).SetInt(AssetInAmount), assetInAmountMinusFee)

	tmp2 := new(big.Float).Add(new(big.Float).SetInt(InputSupply), assetInAmountMinusFee)
	tmp2 = new(big.Float).Quo(new(big.Float).SetInt(k), tmp2)
	assetOutAmount := new(big.Float).Sub(new(big.Float).SetInt(OutputSupply), tmp2)

	assetOutAmountInt, _ := assetOutAmount.Int(nil)

	amountOut := assets.AssetAmount{Asset: assetOut, Amount: assetOutAmountInt.String()}

	quote = &SwapQuote{
		SwapType:  "fixed-input",
		AmountIn:  amountIn,
		AmountOut: &amountOut,
		SwapFees:  &assets.AssetAmount{Asset: amountIn.Asset, Amount: swapFees.String()},
		Slippage:  slippage,
	}

	return

}

func (s *Pool) FetchFixedInputSwapQuoteWithDefaultSlippage(amountIn *assets.AssetAmount) (quote *SwapQuote, err error) {
	return s.FetchFixedInputSwapQuote(amountIn, 0.05)
}

func (s *Pool) FetchFixedOutputSwapQuote(amountOut *assets.AssetAmount, slippage float64) (quote *SwapQuote, err error) {

	var assetIn *assets.Asset
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

	InputSupply, _ := new(big.Float).SetString(inputSupply)
	OutputSupply, _ := new(big.Float).SetString(outputSupply)
	AssetOutAmount, _ := new(big.Float).SetString(assetOutAmount)

	k := new(big.Float).Mul(InputSupply, OutputSupply)

	tmp := new(big.Float).Quo(k, new(big.Float).Sub(OutputSupply, AssetOutAmount))
	calculatedAmountInWithoutFee := new(big.Float).Sub(tmp, InputSupply)

	assetInAmount := new(big.Float).Mul(calculatedAmountInWithoutFee, big.NewFloat(1000))
	assetInAmount.Quo(assetInAmount, big.NewFloat(997))

	swapFees := new(big.Float).Sub(assetInAmount, calculatedAmountInWithoutFee)

	swapFeesInt, _ := swapFees.Int(nil)
	assetInAmountInt, _ := assetInAmount.Int(nil)

	amountIn := assets.AssetAmount{Asset: assetIn, Amount: assetInAmountInt.String()}

	quote = &SwapQuote{
		SwapType:  "fixed-output",
		AmountIn:  &amountIn,
		AmountOut: amountOut,
		SwapFees:  &assets.AssetAmount{Asset: amountIn.Asset, Amount: swapFeesInt.String()},
		Slippage:  slippage,
	}

	return

}

func (s *Pool) FetchFixedOutputSwapQuoteWithDefaultSlippage(amountOut *assets.AssetAmount) (quote *SwapQuote, err error) {
	return s.FetchFixedOutputSwapQuote(amountOut, 0.05)
}

func (s *Pool) PrepareSwapTransactions(amountIn, amountOut *assets.AssetAmount, swapType string, swapperAddress string) (txnGroup *utils.TransactionGroup, err error) {

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
func (s *Pool) PrepareMintTransactions(amountsInStr string, liquidityAssetAmount *assets.AssetAmount, poolerAddress string) (txnGroup *utils.TransactionGroup, err error) {

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

func (s *Pool) PrepareBurnTransactions(liquidityAssetAmount *assets.AssetAmount, amountsOutStr, poolerAddress string) (txnGroup *utils.TransactionGroup, err error) {

	amountsOut := make(map[int]string)
	err = json.Unmarshal([]byte(amountsOutStr), &amountsOut)

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

func (s *Pool) PrepareRedeemTransactions(amountOut *assets.AssetAmount, userAddress string) (txnGroup *utils.TransactionGroup, err error) {

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

	liquidityAssetIn := &assets.AssetAmount{Asset: s.LiquidityAsset, Amount: liquidityAssetAmount}

	quote, err := s.FetchBurnQuoteWithDefaultSlippage(liquidityAssetIn)
	if err != nil {
		return
	}

	LiquidityAssetAmount, _ := new(big.Float).SetString(liquidityAssetAmount)
	IssuedLiquidity, _ := new(big.Float).SetString(s.IssuedLiquidity)

	share := new(big.Float).Quo(LiquidityAssetAmount, IssuedLiquidity)

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

// not compatible with go-mobile
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
