package types

type ASC struct {
	Repo      string `json:"repo"`
	Ref       string `json:"ref"`
	Contracts struct {
		PoolLogicsig struct {
			Type  string          `json:"type"`
			Logic LogicDefinition `json:"logic"`
			Name  string          `json:"name"`
		} `json:"pool_logicsig"`
		ValidatorApp struct {
			Type            string `json:"type"`
			ApprovalProgram struct {
				Bytecode  string        `json:"bytecode"`
				Address   string        `json:"address"`
				Size      int           `json:"size"`
				Variables []interface{} `json:"variables"`
				Source    string        `json:"source"`
			} `json:"approval_program"`
			ClearProgram struct {
				Bytecode  string        `json:"bytecode"`
				Address   string        `json:"address"`
				Size      int           `json:"size"`
				Variables []interface{} `json:"variables"`
				Source    string        `json:"source"`
			} `json:"clear_program"`
			GlobalStateSchema struct {
				NumUints      int `json:"num_uints"`
				NumByteSlices int `json:"num_byte_slices"`
			} `json:"global_state_schema"`
			LocalStateSchema struct {
				NumUints      int `json:"num_uints"`
				NumByteSlices int `json:"num_byte_slices"`
			} `json:"local_state_schema"`
			Name string `json:"name"`
		} `json:"validator_app"`
	} `json:"contracts"`
}

type LogicDefinition struct {
	Bytecode  string `json:"bytecode"`
	Address   string `json:"address"`
	Size      int    `json:"size"`
	Variables []struct {
		Name   string `json:"name"`
		Type   string `json:"type"`
		Index  int    `json:"index"`
		Length int    `json:"length"`
	} `json:"variables"`
	Source string `json:"source"`
}
