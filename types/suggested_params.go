package types

type SuggestedParams struct {
	Fee              int    `json:"fee"`
	GenesisID        string `json:"genesis-id"`
	GenesisHash      []byte `json:"genesis-hash"`
	FirstRoundValid  int    `json:"first-round-valid"`
	LastRoundValid   int    `json:"last-round-valid"`
	ConsensusVersion string `json:"consensus-version"`
	FlatFee          bool   `json:"flat-fee"`
	MinFee           int    `json:"min-fee"`
}
