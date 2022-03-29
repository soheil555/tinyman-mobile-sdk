package assets

import (
	"context"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
)

type Asset struct {
	Id       uint64
	Name     string
	UnitName string
	Decimals uint64
}

//TODO: what about __call__, __hash__, __repr__ methods?
func (s *Asset) Fetch(algod algod.Client) error {

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

//TODO: what about __mul__, __add__, __sub__, __gt__, __lt__, __eq__, __repr__ methods?
type AssetAmount struct {
	Asset  Asset
	Amount int
}
