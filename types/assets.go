package types

import (
	"context"
	"fmt"
	"math"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
)

//TODO: problem on using uint64 or float64
type Asset struct {
	Id       uint64
	Name     string
	UnitName string
	Decimals uint64
}

func (s *Asset) Fetch(algod *algod.Client) error {

	var params models.AssetParams

	if s.Id > 0 {
		asset, err := algod.GetAssetByID(s.Id).Do(context.Background())

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

	return nil

}

//TODO: is call and hash methods ok?
func (s *Asset) Call(amount float64) AssetAmount {
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
	Amount float64
}

func (s *AssetAmount) Mul(o float64) AssetAmount {

	return AssetAmount{s.Asset, s.Amount * o}
}

func (s *AssetAmount) Add(o AssetAmount) (AssetAmount, error) {
	if s.Asset != o.Asset {
		return AssetAmount{}, fmt.Errorf("unsupported asset type for +")
	}

	return AssetAmount{s.Asset, s.Amount + o.Amount}, nil
}

//TODO: can amount be negative?
func (s *AssetAmount) Sub(o AssetAmount) (AssetAmount, error) {
	if s.Asset != o.Asset {
		return AssetAmount{}, fmt.Errorf("unsupported asset type for -")
	}

	return AssetAmount{s.Asset, s.Amount - o.Amount}, nil
}

//TODO: check for int and float
func (s *AssetAmount) Gt(o AssetAmount) (bool, error) {
	if s.Asset != o.Asset {
		return false, fmt.Errorf("unsupported asset type for >")
	}

	return s.Amount > o.Amount, nil
}

//TODO: check for int and float
func (s *AssetAmount) Lt(o AssetAmount) (bool, error) {
	if s.Asset != o.Asset {
		return false, fmt.Errorf("unsupported asset type for <")
	}

	return s.Amount < o.Amount, nil
}

func (s *AssetAmount) String() string {
	amount := float64(s.Amount) / math.Pow(10.0, float64(s.Asset.Decimals))
	return fmt.Sprintf("{%s}('%f')", s.Asset.UnitName, amount)
}
