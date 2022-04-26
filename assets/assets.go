package assets

import (
	"context"
	"fmt"
	"math/big"

	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
	"github.com/algorand/go-algorand-sdk/client/v2/indexer"
	"github.com/soheil555/tinyman-mobile-sdk/utils"
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

	sAmount := utils.NewBigFloatString(s.Amount)
	product := new(big.Float)

	product.Mul(sAmount, big.NewFloat(other))

	productInt, _ := product.Int(nil)
	assetAmount = &AssetAmount{s.Asset, productInt.String()}

	return

}

func (s *AssetAmount) Add(other *AssetAmount) (assetAmount *AssetAmount, err error) {

	if *s.Asset != *other.Asset {
		err = fmt.Errorf("unsupported asset type for +")
		return
	}

	sAmount := utils.NewBigIntString(s.Amount)
	oAmount := utils.NewBigIntString(other.Amount)

	sum := new(big.Int)
	sum.Add(sAmount, oAmount)

	assetAmount = &AssetAmount{s.Asset, sum.String()}

	return

}

func (s *AssetAmount) Sub(other *AssetAmount) (assetAmount *AssetAmount, err error) {

	if *s.Asset != *other.Asset {
		err = fmt.Errorf("unsupported asset type for -")
		return
	}

	sAmount := utils.NewBigIntString(s.Amount)

	oAmount := utils.NewBigIntString(other.Amount)

	difference := new(big.Int)
	difference.Sub(sAmount, oAmount)

	assetAmount = &AssetAmount{s.Asset, difference.String()}
	return

}

func (s *AssetAmount) Eq(other *AssetAmount) (bool, error) {

	if *s.Asset != *other.Asset {
		return false, fmt.Errorf("unsupported asset type for ==")
	}

	sAmount := utils.NewBigIntString(s.Amount)
	oAmount := utils.NewBigIntString(other.Amount)

	return sAmount.Cmp(oAmount) == 0, nil

}

func (s *AssetAmount) Gt(other *AssetAmount) (bool, error) {

	if *s.Asset != *other.Asset {
		return false, fmt.Errorf("unsupported asset type for >")
	}

	sAmount := utils.NewBigIntString(s.Amount)
	oAmount := utils.NewBigIntString(other.Amount)

	return sAmount.Cmp(oAmount) > 0, nil

}

func (s *AssetAmount) Lt(other *AssetAmount) (bool, error) {

	if *s.Asset != *other.Asset {
		return false, fmt.Errorf("unsupported asset type for <")
	}

	sAmount := utils.NewBigIntString(s.Amount)
	oAmount := utils.NewBigIntString(other.Amount)

	return sAmount.Cmp(oAmount) < 0, nil

}

func (s *AssetAmount) String() string {

	sAmount := utils.NewBigFloatString(s.Amount)

	helper := new(big.Int)
	helper.Exp(big.NewInt(10), big.NewInt(int64(s.Asset.Decimals)), nil)

	amount := new(big.Float)
	amount.Quo(sAmount, new(big.Float).SetInt(helper))

	return fmt.Sprintf("%s('%s')", s.Asset.UnitName, amount.String())

}
