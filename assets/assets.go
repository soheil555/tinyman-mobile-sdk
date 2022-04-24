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

func NewAsset(id int, name string, unitName string, decimals int) *Asset {
	return &Asset{id, name, unitName, decimals}
}

// not compatible with go-mobile
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

func NewAssetAmount(asset *Asset, amount string) *AssetAmount {
	return &AssetAmount{asset, amount}
}

func (s *AssetAmount) Mul(other float64) (assetAmount *AssetAmount) {

	sAmount, ok := new(big.Float).SetString(s.Amount)
	if !ok {
		return
	}

	mulResult := new(big.Float).Mul(sAmount, big.NewFloat(other))
	Amount, _ := mulResult.Int(nil)

	assetAmount = &AssetAmount{s.Asset, Amount.String()}
	return

}

func (s *AssetAmount) Add(other *AssetAmount) (assetAmount *AssetAmount, err error) {

	if *s.Asset != *other.Asset {
		err = fmt.Errorf("unsupported asset type for +")
		return
	}

	sAmount, ok := new(big.Int).SetString(s.Amount, 10)
	if !ok {
		return
	}

	oAmount, ok := new(big.Int).SetString(other.Amount, 10)
	if !ok {
		return
	}

	Amount := new(big.Int).Add(sAmount, oAmount)
	assetAmount = &AssetAmount{s.Asset, Amount.String()}

	return

}

func (s *AssetAmount) Sub(other *AssetAmount) (assetAmount *AssetAmount, err error) {

	if *s.Asset != *other.Asset {
		err = fmt.Errorf("unsupported asset type for -")
		return
	}

	sAmount, ok := new(big.Int).SetString(s.Amount, 10)
	if !ok {
		return
	}

	oAmount, ok := new(big.Int).SetString(other.Amount, 10)
	if !ok {
		return
	}

	Amount := new(big.Int).Sub(sAmount, oAmount)

	assetAmount = &AssetAmount{s.Asset, Amount.String()}
	return

}

func (s *AssetAmount) Eq(other *AssetAmount) (bool, error) {

	if *s.Asset != *other.Asset {
		return false, fmt.Errorf("unsupported asset type for ==")
	}

	sAmount, ok := new(big.Int).SetString(s.Amount, 10)
	if !ok {
		return false, nil
	}

	oAmount, ok := new(big.Int).SetString(other.Amount, 10)
	if !ok {
		return false, nil
	}

	return sAmount.Cmp(oAmount) == 0, nil

}

func (s *AssetAmount) Gt(other *AssetAmount) (bool, error) {

	if *s.Asset != *other.Asset {
		return false, fmt.Errorf("unsupported asset type for >")
	}

	sAmount, ok := new(big.Int).SetString(s.Amount, 10)
	if !ok {
		return false, nil
	}

	oAmount, ok := new(big.Int).SetString(other.Amount, 10)
	if !ok {
		return false, nil
	}

	return sAmount.Cmp(oAmount) > 0, nil

}

func (s *AssetAmount) Lt(other *AssetAmount) (bool, error) {

	if *s.Asset != *other.Asset {
		return false, fmt.Errorf("unsupported asset type for <")
	}

	sAmount, ok := new(big.Int).SetString(s.Amount, 10)
	if !ok {
		return false, nil
	}

	oAmount, ok := new(big.Int).SetString(other.Amount, 10)
	if !ok {
		return false, nil
	}

	return sAmount.Cmp(oAmount) < 0, nil

}

func (s *AssetAmount) String() string {

	sAmount, ok := new(big.Float).SetString(s.Amount)
	if !ok {
		return ""
	}
	tmp := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(s.Asset.Decimals)), nil)
	amount := new(big.Float).Quo(sAmount, new(big.Float).SetInt(tmp))
	return fmt.Sprintf("%s('%s')", s.Asset.UnitName, amount.String())

}
