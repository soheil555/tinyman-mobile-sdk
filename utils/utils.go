package utils

import (
	"context"
	"crypto/ed25519"
	b64 "encoding/base64"
	"encoding/binary"
	"fmt"
	"sort"
	"strings"

	"github.com/soheil555/tinyman-mobile-sdk/types"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
	"github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/transaction"
	algoTypes "github.com/algorand/go-algorand-sdk/types"
)

/*
	Return a byte array to be used in LogicSig.
*/
func GetProgram(definition types.Logic, variables map[string]int) (templateBytes []byte, err error) {

	template := definition.Bytecode

	templateBytes, err = b64.StdEncoding.DecodeString(template)
	if err != nil {
		return
	}

	offset := 0
	var dVariables = definition.Variables

	sort.SliceStable(dVariables, func(i, j int) bool {
		return dVariables[i].Index < dVariables[j].Index
	})

	for _, v := range dVariables {

		s := strings.Split(v.Name, "TMPL_")
		name := strings.ToLower(s[len(s)-1])
		value := variables[name]
		start := v.Index - offset
		end := start + v.Length

		var valueEncoded []byte
		valueEncoded, err = EncodeValue(value, v.Type)

		if err != nil {
			return
		}

		valueEncodedLen := len(valueEncoded)
		diff := v.Length - valueEncodedLen
		offset += diff

		var tmp []byte
		tmp = append(tmp, templateBytes[:start]...)
		tmp = append(tmp, valueEncoded...)
		tmp = append(tmp, templateBytes[end:]...)

		templateBytes = tmp

	}

	return

}

func EncodeValue(value int, _type string) (buf []byte, err error) {

	if _type == "int" {

		buf = EncodeVarint(value)
		return

	}

	err = fmt.Errorf("unsupported value type %s", _type)
	return

}

func EncodeVarint(number int) (buf []byte) {

	for {
		towrite := number & 0x7f
		number >>= 7

		if number != 0 {
			buf = append(buf, byte(towrite|0x80))
		} else {
			buf = append(buf, byte(towrite))
			break
		}
	}

	return

}

func SignAndSubmitTransactions(client *algod.Client, transactions []algoTypes.Transaction, signedTransactions [][]byte, sender algoTypes.Address, senderSK ed25519.PrivateKey) (trxInfo *types.TrxInfo, err error) {

	for i, txn := range transactions {

		if txn.Sender == sender {
			var stx []byte
			_, stx, err = crypto.SignTransaction(senderSK, txn)

			if err != nil {
				err = fmt.Errorf("signing failed with %v", err)
				return
			}

			signedTransactions[i] = stx
		}

	}

	var signedGroup []byte

	for _, stx := range signedTransactions {

		signedGroup = append(signedGroup, stx...)

	}

	txid, err := client.SendRawTransaction(signedGroup).Do(context.Background())

	if err != nil {
		err = fmt.Errorf("failed to create transaction: %v", err)
		return
	}

	return WaitForConfirmation(client, txid)

}

/*
   Utility function to wait until the transaction is
   confirmed before proceeding.
*/
func WaitForConfirmation(client *algod.Client, txid string) (trxInfo *types.TrxInfo, err error) {

	nodeStatus, err := client.Status().Do(context.Background())
	if err != nil {
		err = fmt.Errorf("error getting algod status: %s", err)
		return
	}

	lastRound := nodeStatus.LastRound

	txinfo := client.PendingTransactionInformation(txid)

	pendingTrxInfo, _, err := txinfo.Do(context.Background())

	if err != nil {
		err = fmt.Errorf("error getting algod pending transaction info: %s", err)
		return
	}

	for !(pendingTrxInfo.ConfirmedRound > 0) {

		fmt.Println("Waiting for confirmation")
		lastRound += 1

		_, err = client.StatusAfterBlock(lastRound).Do(context.Background())
		if err != nil {
			err = fmt.Errorf("error getting status after block: %s", err)
			return
		}

		pendingTrxInfo, _, err = txinfo.Do(context.Background())

		if err != nil {
			err = fmt.Errorf("error getting algod pending transaction info: %s", err)
			return
		}

	}

	fmt.Printf("Transaction %s confirmed in round %d.\n", txid, pendingTrxInfo.ConfirmedRound)

	trxInfo = &types.TrxInfo{
		TxId:           txid,
		ConfirmedRound: int(pendingTrxInfo.ConfirmedRound),
	}

	return

}

func IntToBytes(num int) []byte {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data[:], uint64(num))
	return data
}

func GetStateInt(state map[string]models.TealValue, key interface{}) int {

	var keyString string

	switch k := key.(type) {

	case string:
		keyString = b64.StdEncoding.EncodeToString([]byte(k))
	case []byte:
		keyString = string(k[:])
	default:
		keyString = ""

	}

	if val, ok := state[keyString]; ok {
		return int(val.Uint)
	}
	return 0

}

func GetStateBytes(state map[string]models.TealValue, key interface{}) string {

	var keyString string

	switch k := key.(type) {

	case string:
		keyString = b64.StdEncoding.EncodeToString([]byte(k))
	case []byte:
		keyString = string(k[:])
	default:
		keyString = ""

	}

	if val, ok := state[keyString]; ok {
		return val.Bytes
	}
	return ""

}

type TransactionGroup struct {
	transactions       []algoTypes.Transaction
	signedTransactions [][]byte
}

func NewTransactionGroup(transactions []algoTypes.Transaction) (transactionGroup *TransactionGroup, err error) {

	transactions, err = transaction.AssignGroupID(transactions, "")
	if err != nil {
		return
	}

	signedTransactions := make([][]byte, len(transactions))
	return &TransactionGroup{transactions, signedTransactions}, nil

}

func (s *TransactionGroup) GetTransactions() []algoTypes.Transaction {
	return s.transactions
}

func (s *TransactionGroup) GetSignedTransactions() [][]byte {
	return s.signedTransactions
}

// TODO: what is user
// type User interface {
// 	SignTransactionGroup(transactionGroup *TransactionGroup)
// }

// func (s *TransactionGroup) Sign(user User) {
// 	user.SignTransactionGroup(s)
// }

func (s *TransactionGroup) GetSignedGroup() (signedGroup []byte) {

	for _, txn := range s.signedTransactions {
		signedGroup = append(signedGroup, txn...)
	}

	return

}

func (s *TransactionGroup) SignWithLogicsig(logicsig types.LogicSig) (err error) {

	lsig := algoTypes.LogicSig{
		Logic: logicsig.Logic,
	}

	address := crypto.AddressFromProgram(logicsig.Logic)

	for i, txn := range s.transactions {
		if txn.Sender == address {
			_, stxBytes, err := crypto.SignLogicsigTransaction(lsig, txn)

			if err != nil {
				return fmt.Errorf("failed to sign transaction: %v", err)
			}

			s.signedTransactions[i] = stxBytes
		}
	}

	return

}

func (s *TransactionGroup) SignWithPrivateKey(address string, privateKey string) (err error) {

	for i, txn := range s.transactions {
		if txn.Sender.String() == address {
			_, stxBytes, err := crypto.SignTransaction([]byte(privateKey), txn)
			if err != nil {
				return fmt.Errorf("failed to sign transaction: %v", err)
			}
			s.signedTransactions[i] = stxBytes
		}
	}

	return

}

func (s *TransactionGroup) Sumbit(algod *algod.Client, wait bool) (trxInfo *types.TrxInfo, err error) {

	var signedGroup []byte

	for _, txn := range s.signedTransactions {

		signedGroup = append(signedGroup, txn...)

	}

	txid, err := algod.SendRawTransaction(signedGroup).Do(context.Background())

	if err != nil {
		err = fmt.Errorf("failed to send transaction: %v", err)
		return
	}

	if wait {
		return WaitForConfirmation(algod, txid)
	}

	trxInfo = &types.TrxInfo{
		TxId: txid,
	}
	return

}
