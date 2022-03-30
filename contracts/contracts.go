package contracts

import (
	"encoding/json"
	"os"
	"sort"
	myTypes "tinyman-mobile-sdk/types"
	"tinyman-mobile-sdk/utils"

	"github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/types"
)

func readContractsFile(fileName string) (myTypes.ASC, error) {

	datas := myTypes.ASC{}

	file, err := os.ReadFile(fileName)

	if err != nil {
		return myTypes.ASC{}, err
	}

	json.Unmarshal(file, &datas)

	return datas, nil

}

func GetPoolLogicsig(validatorAppID int, asset1ID int, asset2ID int) (types.LogicSig, error) {

	contracts, err := readContractsFile("../asc.json")

	if err != nil {
		return types.LogicSig{}, err
	}

	poolLogicsigDef := contracts.Contracts.PoolLogicsig.Logic
	// validatorAppDef := contracts.Contracts.ValidatorApp

	assets := []int{asset1ID, asset2ID}
	sort.Ints(assets)

	assetID1 := assets[1]
	assetID2 := assets[0]

	variables := map[string]interface{}{
		"validator_app_id": validatorAppID,
		"asset_id_1":       assetID1,
		"asset_id_2":       assetID2,
	}

	programBytes, err := utils.GetProgram(poolLogicsigDef, variables)
	if err != nil {
		return types.LogicSig{}, err
	}

	var args [][]byte

	ma := crypto.MultisigAccount{}

	lsig, err := crypto.MakeLogicSig(programBytes, args, nil, ma)

	if err != nil {
		return types.LogicSig{}, err
	}

	return lsig, nil

}
