package assets

import (
	"context"
	"fmt"
	"math/big"

	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
	"github.com/algorand/go-algorand-sdk/client/v2/indexer"
)

type Asset struct {
	Id       int
	Name     string
	UnitName string
	Decimals int
}

func (s *Asset) Fetch(indexer *indexer.Client) (err error) {

	var params models.AssetParams

	if s.Id > 0 {

		_, asset, err := indexer.LookupAssetByID(uint64(s.Id)).Do(context.Background())

		if err != nil {
			return err
		}

		params = asset.Params

	} else {

		params = models.AssetParams{
			Name:     "Algo",
			UnitName: "ALGO",
			Decimals: 6,
		}

	}

	s.Name = params.Name
	s.UnitName = params.UnitName
	s.Decimals = int(params.Decimals)

	return

}

func (s *Asset) Call(amount string) (assetAmount *AssetAmount) {
	return &AssetAmount{s, amount}
}

func (s *Asset) Hash() int {
	return s.Id
}

func (s *Asset) String() string {
	return fmt.Sprintf("Asset(%s - %d)", s.UnitName, s.Id)
}

type AssetAmount struct {
	Asset  *Asset
	Amount string
}

func (s *AssetAmount) Mul(o float64) (assetAmount *AssetAmount) {

	sAmount, ok := new(big.Float).SetString(s.Amount)
	if !ok {
		return
	}

	mulResult := new(big.Float).Mul(sAmount, big.NewFloat(o))
	Amount, _ := mulResult.Int(nil)

	assetAmount = &AssetAmount{s.Asset, Amount.String()}
	return

}

func (s *AssetAmount) Add(o *AssetAmount) (assetAmount *AssetAmount, err error) {

	if s.Asset != o.Asset {
		err = fmt.Errorf("unsupported asset type for +")
		return
	}

	sAmount, ok := new(big.Int).SetString(s.Amount, 10)
	if !ok {
		return
	}

	oAmount, ok := new(big.Int).SetString(o.Amount, 10)
	if !ok {
		return
	}

	Amount := new(big.Int).Add(sAmount, oAmount)
	assetAmount = &AssetAmount{s.Asset, Amount.String()}

	return

}

func (s *AssetAmount) Sub(o *AssetAmount) (assetAmount *AssetAmount, err error) {

	if s.Asset != o.Asset {
		err = fmt.Errorf("unsupported asset type for -")
		return
	}

	sAmount, ok := new(big.Int).SetString(s.Amount, 10)
	if !ok {
		return
	}

	oAmount, ok := new(big.Int).SetString(o.Amount, 10)
	if !ok {
		return
	}

	Amount := new(big.Int).Add(sAmount, oAmount)

	assetAmount = &AssetAmount{s.Asset, Amount.String()}
	return

}

func (s *AssetAmount) Eq(o *AssetAmount) (bool, error) {

	if s.Asset != o.Asset {
		return false, fmt.Errorf("unsupported asset type for ==")
	}

	sAmount, ok := new(big.Int).SetString(s.Amount, 10)
	if !ok {
		return false, nil
	}

	oAmount, ok := new(big.Int).SetString(o.Amount, 10)
	if !ok {
		return false, nil
	}

	return sAmount.Cmp(oAmount) == 0, nil

}

func (s *AssetAmount) Gt(o *AssetAmount) (bool, error) {

	if s.Asset != o.Asset {
		return false, fmt.Errorf("unsupported asset type for >")
	}

	sAmount, ok := new(big.Int).SetString(s.Amount, 10)
	if !ok {
		return false, nil
	}

	oAmount, ok := new(big.Int).SetString(o.Amount, 10)
	if !ok {
		return false, nil
	}

	return sAmount.Cmp(oAmount) > 0, nil

}

func (s *AssetAmount) Lt(o *AssetAmount) (bool, error) {

	if s.Asset != o.Asset {
		return false, fmt.Errorf("unsupported asset type for <")
	}

	sAmount, ok := new(big.Int).SetString(s.Amount, 10)
	if !ok {
		return false, nil
	}

	oAmount, ok := new(big.Int).SetString(o.Amount, 10)
	if !ok {
		return false, nil
	}

	return sAmount.Cmp(oAmount) < 0, nil

}

func (s *AssetAmount) String() string {

	sAmount, ok := new(big.Int).SetString(s.Amount, 10)
	if !ok {
		return ""
	}

	amount := new(big.Int).Div(sAmount, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(s.Asset.Decimals)), nil))
	return fmt.Sprintf("%s('%s')", s.Asset.UnitName, amount.String())

}
