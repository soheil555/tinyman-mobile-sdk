package types

type TransactionInformation struct {
	TxId           string `json:"tx-id"`
	ConfirmedRound int    `json:"confirmed-round"`
}
