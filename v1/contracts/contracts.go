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

	json.Unmarshal(file, &data)

	return

}

func GetPoolLogicsig(validatorAppID int, asset1ID int, asset2ID int) (lsig types.LogicSig, err error) {

	contracts, err := readContractsFile()

	if err != nil {
		return
	}

	poolLogicsigDefBytes, err := json.Marshal(contracts.Contracts.PoolLogicsig.Logic)
	if err != nil {
		return
	}
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

	variablesBytes, err := json.Marshal(variables)
	if err != nil {
		return
	}

	programBytes, err := utils.GetProgram(poolLogicsigDefBytes, variablesBytes)
	if err != nil {
		return
	}

	logsicSig, err := crypto.MakeLogicSig(programBytes, [][]byte{}, nil, crypto.MultisigAccount{})
	if err != nil {
		return
	}

	lsig.Logic = logsicSig.Logic
	lsig.Msig = logsicSig.Msig
	lsig.Sig = logsicSig.Sig

	return

}
