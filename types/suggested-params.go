package types

type SuggestedParams struct {
	Fee              int
	GenesisID        string
	GenesisHash      []byte
	FirstRoundValid  int
	LastRoundValid   int
	ConsensusVersion string
	FlatFee          bool
	MinFee           int
}
