package contracts

import (
	"embed"
	"encoding/json"
	"sort"
	"tinyman-mobile-sdk/types"
	"tinyman-mobile-sdk/utils"

	"github.com/algorand/go-algorand-sdk/crypto"
	algoTypes "github.com/algorand/go-algorand-sdk/types"
)

//go:embed asc.json
var f embed.FS

func readContractsFile() (data types.ASC, err error) {

	file, err := f.ReadFile("asc.json")

	if err != nil {
		return
	}

	json.Unmarshal(file, &data)

	return

}

func GetPoolLogicsig(validatorAppID uint64, asset1ID uint64, asset2ID uint64) (lsig algoTypes.LogicSig, err error) {

	contracts, err := readContractsFile()

	if err != nil {
		return
	}

	poolLogicsigDef := contracts.Contracts.PoolLogicsig.Logic
	// validatorAppDef := contracts.Contracts.ValidatorApp

	assets := []uint64{asset1ID, asset2ID}
	sort.Slice(assets, func(i, j int) bool { return assets[i] < assets[j] })

	assetID1 := assets[1]
	assetID2 := assets[0]

	variables := map[string]uint64{
		"validator_app_id": validatorAppID,
		"asset_id_1":       assetID1,
		"asset_id_2":       assetID2,
	}

	programBytes, err := utils.GetProgram(poolLogicsigDef, variables)
	if err != nil {
		return
	}

	lsig, err = crypto.MakeLogicSig(programBytes, [][]byte{}, nil, crypto.MultisigAccount{})

	return

}
