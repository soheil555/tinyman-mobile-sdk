package utils

import (
	"context"
	"crypto/ed25519"
	b64 "encoding/base64"
	"encoding/binary"
	"fmt"
	"sort"
	"strings"
	"tinyman-mobile-sdk/types"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
	"github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/transaction"
	algoTypes "github.com/algorand/go-algorand-sdk/types"
)

/*
	Return a byte array to be used in LogicSig.
*/
func GetProgram(definition types.LogicDefinition, variables map[string]interface{}) (templateBytes []byte, err error) {

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

		//TODO: better way maybe
		var valueEncoded []byte
		valueEncoded, err = EncodeValue(value, v.Type)

		if err != nil {
			return
		}

		valueEncodedLen := len(valueEncoded)
		diff := v.Length - valueEncodedLen
		offset += diff

		//TODO: better way for assign?
		var tmp []byte
		tmp = append(tmp, templateBytes[:start]...)
		tmp = append(tmp, valueEncoded...)
		tmp = append(tmp, templateBytes[end:]...)

		templateBytes = tmp

	}

	return

}

func EncodeValue(value interface{}, _type string) (buf []byte, err error) {

	if _type == "int" {

		var uintValue uint64

		//TODO: better way maybe
		switch v := value.(type) {

		case int:
			uintValue = uint64(v)
		case uint64:
			uintValue = v
		default:
			return nil, fmt.Errorf("only int or uint64 type is valid for value")

		}

		buf = EncodeVarint(uintValue)
		return

	} else {

		err = fmt.Errorf("unsupported value type %s", _type)
		return

	}

}

func EncodeVarint(number uint64) (buf []byte) {

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

func SignAndSubmitTransactions(client *algod.Client, transactions []algoTypes.Transaction, signedTransactions [][]byte, sender algoTypes.Address, senderSK ed25519.PrivateKey) (pendingTrxInfo models.PendingTransactionInfoResponse, Txid string, err error) {

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
func WaitForConfirmation(client *algod.Client, txid string) (pendingTrxInfo models.PendingTransactionInfoResponse, Txid string, err error) {

	nodeStatus, err := client.Status().Do(context.Background())
	if err != nil {
		err = fmt.Errorf("error getting algod status: %s", err)
		return
	}

	lastRound := nodeStatus.LastRound

	txinfo := client.PendingTransactionInformation(txid)

	pendingTrxInfo, _, err = txinfo.Do(context.Background())

	if err != nil {
		err = fmt.Errorf("error getting algod pending transaction info: %s", err)
		return
	}

	//TODO: is it correct?
	for !(pendingTrxInfo.ConfirmedRound > 0) {

		fmt.Println("Waiting for confirmation")
		lastRound += 1
		//TODO: do nothing with response?
		_, err = client.StatusAfterBlock(lastRound).Do(context.Background())
		if err != nil {
			err = fmt.Errorf("error getting algod pending transaction info: %s", err)
			return
		}

		pendingTrxInfo, _, err = txinfo.Do(context.Background())

		if err != nil {
			err = fmt.Errorf("error getting algod pending transaction info: %s", err)
			return
		}

	}

	//TODO: is it good way to return txid
	fmt.Printf("Transaction %s confirmed in round %d.\n", txid, pendingTrxInfo.ConfirmedRound)
	return pendingTrxInfo, txid, nil

}

func IntToBytes(num uint64) []byte {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data[:], num)
	return data
}

func GetStateInt(state map[string]models.TealValue, key interface{}) uint64 {

	var keyString string

	switch k := key.(type) {

	case string:
		keyString = b64.StdEncoding.EncodeToString([]byte(k))
	case []byte:
		keyString = string(k[:])
	default:
		//TODO: maybe return error
		keyString = ""

	}

	if val, ok := state[keyString]; ok {
		return val.Uint
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
		//TODO: maybe return error
		keyString = ""

	}

	if val, ok := state[keyString]; ok {
		return val.Bytes
	}
	return ""

}

//TODO: should move to another file?
type TransactionGroup struct {
	Transactions       []algoTypes.Transaction
	SignedTransactions [][]byte
}

func NewTransactionGroup(transactions []algoTypes.Transaction) (transactionGroup TransactionGroup, err error) {

	transactions, err = transaction.AssignGroupID(transactions, "")
	if err != nil {
		return TransactionGroup{}, err
	}
	//TODO: [][]byte. is it good?
	signedTransactions := make([][]byte, len(transactions))
	return TransactionGroup{transactions, signedTransactions}, nil

}

// TODO: what is user
type User interface {
	SignTransactionGroup(transactionGroup *TransactionGroup)
}

func (s *TransactionGroup) Sign(user User) {
	user.SignTransactionGroup(s)
}

func (s *TransactionGroup) SignWithLogicsig(logicsig algoTypes.LogicSig) (err error) {

	address := crypto.AddressFromProgram(logicsig.Logic)

	for i, txn := range s.Transactions {
		if txn.Sender == address {
			_, stxBytes, err := crypto.SignLogicsigTransaction(logicsig, txn)

			if err != nil {
				return fmt.Errorf("failed to create transaction: %v", err)
			}

			s.SignedTransactions[i] = stxBytes
		}
	}

	return

}

func (s *TransactionGroup) SignWithPrivateKey(address algoTypes.Address, privateKey ed25519.PrivateKey) (err error) {

	for i, txn := range s.Transactions {
		if txn.Sender == address {
			_, stxBytes, err := crypto.SignTransaction(privateKey, txn)
			if err != nil {
				return fmt.Errorf("failed to sign transaction: %v", err)
			}
			s.SignedTransactions[i] = stxBytes
		}
	}

	return

}

func (s *TransactionGroup) Sumbit(algod *algod.Client, wait bool) (pendingTrxInfo models.PendingTransactionInfoResponse, Txid string, err error) {

	var signedGroup []byte

	for _, txn := range s.SignedTransactions {

		signedGroup = append(signedGroup, txn...)

	}

	Txid, err = algod.SendRawTransaction(signedGroup).Do(context.Background())

	if err != nil {
		err = fmt.Errorf("failed to send transaction: %v", err)
		return
	}

	if wait {
		return WaitForConfirmation(algod, Txid)
	}

	return

}
