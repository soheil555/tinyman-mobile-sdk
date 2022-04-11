package types

type ASC struct {
	Repo      string    `json:"repo"`
	Ref       string    `json:"ref"`
	Contracts Contracts `json:"contracts"`
}
type Variable struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Index  int    `json:"index"`
	Length int    `json:"length"`
}
type Logic struct {
	Bytecode  string     `json:"bytecode"`
	Address   string     `json:"address"`
	Size      int        `json:"size"`
	Variables []Variable `json:"variables"`
	Source    string     `json:"source"`
}
type PoolLogicsig struct {
	Type  string `json:"type"`
	Logic Logic  `json:"logic"`
	Name  string `json:"name"`
}
type ApprovalProgram struct {
	Bytecode  string        `json:"bytecode"`
	Address   string        `json:"address"`
	Size      int           `json:"size"`
	Variables []interface{} `json:"variables"`
	Source    string        `json:"source"`
}
type ClearProgram struct {
	Bytecode  string        `json:"bytecode"`
	Address   string        `json:"address"`
	Size      int           `json:"size"`
	Variables []interface{} `json:"variables"`
	Source    string        `json:"source"`
}
type GlobalStateSchema struct {
	NumUints      int `json:"num_uints"`
	NumByteSlices int `json:"num_byte_slices"`
}
type LocalStateSchema struct {
	NumUints      int `json:"num_uints"`
	NumByteSlices int `json:"num_byte_slices"`
}
type ValidatorApp struct {
	Type              string            `json:"type"`
	ApprovalProgram   ApprovalProgram   `json:"approval_program"`
	ClearProgram      ClearProgram      `json:"clear_program"`
	GlobalStateSchema GlobalStateSchema `json:"global_state_schema"`
	LocalStateSchema  LocalStateSchema  `json:"local_state_schema"`
	Name              string            `json:"name"`
}
type Contracts struct {
	PoolLogicsig PoolLogicsig `json:"pool_logicsig"`
	ValidatorApp ValidatorApp `json:"validator_app"`
}
