package types

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/algorand/go-algorand-sdk/client/v2/indexer"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestNewAsset(t *testing.T) {

	expected := Asset{
		1,
		"Test",
		"TEST",
		6,
	}
	result := NewAsset(1, "Test", "TEST", 6)

	assert.Equal(t, expected, *result)

}

func TestAssetFetch(t *testing.T) {

	defer gock.Off()

	mockServerURL := "https://mockserver.com"
	assetID := 1

	gock.New(mockServerURL).Get(fmt.Sprintf("/v2/assets/%d", assetID)).
		Reply(200).JSON(map[string]map[string]map[string]interface{}{
		"asset": {
			"params": {
				"decimals":  6,
				"name":      "Test",
				"unit-name": "TEST",
			},
		},
	})

	asset := &Asset{assetID, "", "", 0}
	indexer, err := indexer.MakeClient(mockServerURL, "")
	assert.Nil(t, err)

	err = asset.Fetch(indexer)
	assert.Nil(t, err)
	expected := Asset{
		assetID,
		"Test",
		"TEST",
		6,
	}

	assert.Equal(t, expected, *asset)

	asset = &Asset{0, "", "", 0}

	err = asset.Fetch(indexer)
	assert.Nil(t, err)
	expected = Asset{
		0,
		"Algo",
		"ALGO",
		6,
	}

	assert.Equal(t, expected, *asset)

}

func TestAssetCall(t *testing.T) {

	asset := &Asset{1, "Test", "TEST", 6}
	assetAmount := asset.Call("1000")

	expected := &AssetAmount{asset, "1000"}

	assert.True(t, reflect.DeepEqual(assetAmount, expected))

}

func TestNewAssetAmount(t *testing.T) {

	asset := &Asset{1, "Test", "TEST", 6}
	assetAmount := NewAssetAmount(asset, "1000")

	expected := &AssetAmount{asset, "1000"}
	assert.True(t, reflect.DeepEqual(assetAmount, expected))

}

func TestAssetAmountMul(t *testing.T) {

	asset := &Asset{1, "Test", "TEST", 6}
	a1 := NewAssetAmount(asset, "1000")

	a2 := a1.Mul(2.5)
	assert.Equal(t, "2500", a2.Amount)

	a1.Amount = "1222"
	a2 = a1.Mul(2.3)
	assert.Equal(t, "2810", a2.Amount)

}

func TestAssetAmountAdd(t *testing.T) {

	asset1 := &Asset{1, "Test", "TEST", 6}
	asset2 := &Asset{2, "Test2", "TEST2", 8}

	a1 := NewAssetAmount(asset1, "1000")
	a2 := NewAssetAmount(asset1, "2000")
	a3 := NewAssetAmount(asset2, "4000")

	result, err := a1.Add(a2)
	assert.Nil(t, err)
	assert.Equal(t, "3000", result.Amount)

	_, err = a1.Add(a3)
	assert.NotNil(t, err)

}

func TestAssetAmountSub(t *testing.T) {

	asset1 := &Asset{1, "Test", "TEST", 6}
	asset2 := &Asset{2, "Test2", "TEST2", 8}

	a1 := NewAssetAmount(asset1, "1000")
	a2 := NewAssetAmount(asset1, "2000")
	a3 := NewAssetAmount(asset2, "4000")

	result, err := a2.Sub(a1)
	assert.Nil(t, err)
	assert.Equal(t, "1000", result.Amount)

	_, err = a1.Sub(a3)
	assert.NotNil(t, err)

}

func TestAssetAmountEq(t *testing.T) {

	asset1 := &Asset{1, "Test", "TEST", 6}
	asset2 := &Asset{2, "Test2", "TEST2", 8}

	a1 := NewAssetAmount(asset1, "2000")
	a2 := NewAssetAmount(asset1, "2000")
	a3 := NewAssetAmount(asset2, "4000")

	result, err := a1.Eq(a2)
	assert.Nil(t, err)
	assert.True(t, result)

	_, err = a1.Eq(a3)
	assert.NotNil(t, err)

}

func TestAssetAmountGt(t *testing.T) {

	asset1 := &Asset{1, "Test", "TEST", 6}
	asset2 := &Asset{2, "Test2", "TEST2", 8}

	a1 := NewAssetAmount(asset1, "3000")
	a2 := NewAssetAmount(asset1, "2000")
	a3 := NewAssetAmount(asset2, "4000")

	result, err := a1.Gt(a2)
	assert.Nil(t, err)
	assert.True(t, result)

	_, err = a1.Gt(a3)
	assert.NotNil(t, err)

}

func TestAssetAmountLt(t *testing.T) {

	asset1 := &Asset{1, "Test", "TEST", 6}
	asset2 := &Asset{2, "Test2", "TEST2", 8}

	a1 := NewAssetAmount(asset1, "2000")
	a2 := NewAssetAmount(asset1, "3000")
	a3 := NewAssetAmount(asset2, "4000")

	result, err := a1.Lt(a2)
	assert.Nil(t, err)
	assert.True(t, result)

	_, err = a1.Lt(a3)
	assert.NotNil(t, err)

}
