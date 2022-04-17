package types

import (
	"context"
	"fmt"
	"math"

	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
	"github.com/algorand/go-algorand-sdk/client/v2/indexer"
)

type Asset struct {
	Id       uint64
	Name     string
	UnitName string
	Decimals uint64
}

func (s *Asset) Fetch(indexer *indexer.Client) (err error) {

	var params models.AssetParams

	if s.Id > 0 {

		_, asset, err := indexer.LookupAssetByID(s.Id).Do(context.Background())

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
	s.Decimals = params.Decimals

	return

}

func (s *Asset) Call(amount uint64) (assetAmount AssetAmount) {
	return AssetAmount{*s, amount}
}

func (s *Asset) Hash() uint64 {
	return s.Id
}

func (s *Asset) String() string {
	return fmt.Sprintf("Asset(%s - %d)", s.UnitName, s.Id)
}

type AssetAmount struct {
	Asset  Asset
	Amount uint64
}

func (s *AssetAmount) Mul(o float64) (assetAmount AssetAmount) {
	return AssetAmount{s.Asset, uint64(float64(s.Amount) * o)}
}

func (s *AssetAmount) Add(o AssetAmount) (assetAmount AssetAmount, err error) {
	if s.Asset != o.Asset {
		err = fmt.Errorf("unsupported asset type for +")
		return
	}

	assetAmount = AssetAmount{s.Asset, s.Amount + o.Amount}

	return
}

//TODO: maybe using an overflow util to handle
func (s *AssetAmount) Sub(o AssetAmount) (assetAmount AssetAmount, err error) {
	if s.Asset != o.Asset {
		err = fmt.Errorf("unsupported asset type for -")
		return
	}

	assetAmount = AssetAmount{s.Asset, s.Amount - o.Amount}
	return
}

func (s *AssetAmount) Gt(o AssetAmount) (bool, error) {
	if s.Asset != o.Asset {
		return false, fmt.Errorf("unsupported asset type for >")
	}

	return s.Amount > o.Amount, nil
}

func (s *AssetAmount) Lt(o AssetAmount) (bool, error) {
	if s.Asset != o.Asset {
		return false, fmt.Errorf("unsupported asset type for <")
	}

	return s.Amount < o.Amount, nil
}

func (s *AssetAmount) String() string {
	amount := float64(s.Amount) / math.Pow(10.0, float64(s.Asset.Decimals))
	return fmt.Sprintf("%s('%f')", s.Asset.UnitName, amount)
}
