package contracts

import (
	"embed"
	"encoding/json"
	"sort"
	"tinyman-mobile-sdk/types"
	"tinyman-mobile-sdk/utils"

	"github.com/algorand/go-algorand-sdk/crypto"
)

//go:embed asc.json
var f embed.FS

func readContractsFile() (data types.ASC, err error) {

	file, err := f.ReadFile("asc.json")

	if err != nil {
		return
	}

	err = json.Unmarshal(file, &data)

	return

}

func GetPoolLogicsig(validatorAppID, asset1ID, asset2ID int) (lsig *types.LogicSig, err error) {

	contracts, err := readContractsFile()

	if err != nil {
		return
	}

	poolLogicsigDef := contracts.Contracts.PoolLogicsig.Logic
	// validatorAppDef := contracts.Contracts.ValidatorApp

	assets := []int{asset1ID, asset2ID}
	sort.Slice(assets, func(i, j int) bool { return assets[i] < assets[j] })

	assetID1 := assets[1]
	assetID2 := assets[0]

	variables := map[string]int{
		"validator_app_id": validatorAppID,
		"asset_id_1":       assetID1,
		"asset_id_2":       assetID2,
	}

	programBytes, err := utils.GetProgram(poolLogicsigDef, variables)
	if err != nil {
		return
	}

	logsicSig, err := crypto.MakeLogicSig(programBytes, [][]byte{}, nil, crypto.MultisigAccount{})
	if err != nil {
		return
	}

	lsig.Logic = logsicSig.Logic

	return

}
