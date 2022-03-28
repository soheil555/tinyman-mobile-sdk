package utils

import (
	b64 "encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/algorand/go-algorand-sdk/transaction"
	"github.com/algorand/go-algorand-sdk/types"
)

type variable struct {
	Name   string
	Type   string
	Index  int
	Length int
}

//TODO: error handling for ,_
//TODO: definition type?
//TODO: definition['variables'] type?
//TODO: is return value type ok?
/*
	Return a byte array to be used in LogicSig.
*/
func GetProgram(definition map[string]interface{}, variables map[string]interface{}) ([]byte, error) {

	template, ok := definition["bytecode"].(string)

	templateBytes, err := b64.StdEncoding.DecodeString(template)
	if err != nil {
		return nil, err
	}

	offset := 0
	var dVariables []variable = definition["variables"].([]variable)

	sort.SliceStable(dVariables, func(i, j int) bool {
		return dVariables[i].Index < dVariables[j].Index
	})

	for _, v := range dVariables {

		s := strings.Split(v.Name, "TMPL_")
		name := strings.ToLower(s[len(s)-1])
		value := variables[name]
		start := v.Index - offset
		end := start + v.Length
		valueEncoded, _ := EncodeValue(value, v.Type)
		valueEncodedLen := len(valueEncoded)
		diff := v.Length - valueEncodedLen
		offset += diff
		//TODO: better way for assign?
		for i := start; i < end; i++ {
			templateBytes[i] = valueEncoded[i-start]
		}

	}

	return templateBytes, nil

}

//TODO: what about type check?
//TODO: what about type as var name?
func EncodeValue(value interface{}, _type string) ([]byte, error) {

	if _type == "int" {
		return EncodeVarint(value.(int)), nil
	} else {
		return nil, errors.New(fmt.Sprintf("Unsupported value type %s!", _type))
	}

}

func EncodeVarint(number int) []byte {

	buf := []byte{}

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

	return buf

}

//TODO: fix return type
//TODO: fix params types
func SignAndSubmitTransactions(client interface{}, transactions []types.Transaction, signedTransactions []types.Transaction, sender types.Address, senderSK interface{}) {

	for i, txn := range transactions {

		if txn.Sender == sender {
			signedTransactions[i] = txn.sign(senderSK)
		}

	}

	txid := client.(signedTransactions)
	return WaitForConfirmation(client, txid)

}

/*
   Utility function to wait until the transaction is
   confirmed before proceeding.
*/

//TODO: fix return type
//TODO: fix params types
func WaitForConfirmation(client interface{}, txid int) {

	lastRound := client.status().get("last-round")
	txinfo := client.pendingTransactionInfo(txid)

	for !(txinfo.get("confirmed-round") && txinfo.get("confirmed-round") > 0) {

		fmt.Println("Waiting for confirmation")
		lastRound += 1
		client.statusAfterBlock(lastRound)
		txinfo = client.pendingTransactionInfo(txid)

	}

	txinfo.txid = txid

	fmt.Printf("Transaction %d confirmed in round %d.\n", txid, txinfo.get("confirmed-round"))
	return txinfo

}

func IntToBytes(num uint) []byte {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data[:], uint64(num))
	return data
}

//TODO: return value type
//TODO: fix bugs
func GetStateInt(state interface{}, key interface{}) {

	switch k := key.(type) {

	case string:
		key = b64.StdEncoding.EncodeToString([]byte(k))

	}

	//TODO: what about decode?
	// return state.get(key, {"unit": 0}).unit

}

//TODO: return value type
//TODO: fix bugs
func GetStateBytes(state interface{}, key interface{}) {

	switch k := key.(type) {

	case string:
		key = b64.StdEncoding.EncodeToString([]byte(k))

	}

	//TODO: what about decode?
	// return state.get(key, {"bytes": ""}).bytes

}

type transactionGroup struct {
	transactions       []types.Transaction
	signedTransactions []types.Transaction
}

func NewTransactionGroup(transactions []types.Transaction) (transactionGroup, error) {

	transactions, err := transaction.AssignGroupID(transactions, "")
	if err != nil {
		return transactionGroup{}, err
	}
	signedTransactions := make([]types.Transaction, len(transactions))
	return transactionGroup{transactions, signedTransactions}, nil

}

//TODO: fix user type
func (s *transactionGroup) Sign(user interface{}) {
	user.signTransactionGroup(s)
}

//TODO: fix logicsig type
//TODO: where is the LogicSigTransaction method
func (s *transactionGroup) SignWithLogicsig(logicsig interface{}) {

	address := logicsig.address()
	for i, txn := range s.transactions {
		if txn.sender == address {
			s.signedTransactions[i] = transaction.LogicSigTransaction(txn, logicsig)
		}
	}

}

func (s *transactionGroup) SignWithPrivateKey(address string, privateKey string) {

	for i, txn := range s.transactions {
		if txn.sender == address {
			self.signedTransactions[i] = txn.sign(privateKey)
		}
	}

}

type Return struct {
	txid interface{}
}

//TODO: fix return type
//TODO: fix algod type errors
//TODO: fix return type for waitforconfirmation
func (s *transactionGroup) Sumbit(algod interface{}, wait bool) (*Return, error) {

	txid, err := algod.sendTransactions(s.signedTransactions)
	if err != nil {
		//TODO: should i return err or new err with string?
		return nil, err
	}

	if wait {
		return WaitForConfirmation(algod, txid)
	}

	return &Return{
		txid,
	}, nil

}
