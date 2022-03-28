package utils

import (
	b64 "encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"sort"
	"strings"
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
func GetProgram(definition map[string]interface{}, variables map[string]interface{}) []byte {

	template, _ := definition["bytecode"].(string)
	templateBytes, _ := b64.StdEncoding.DecodeString(template)

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

	return templateBytes

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

		//TODO: if number == if number != 0 ?
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
func SignAndSubmitTransactions(client interface{}, transactions []interface{}, signedTransactions []interface{}, sender interface{}, senderSK interface{}) {

	for i, txn := range transactions {

		if txn.sender == sender {
			signedTransactions[i] = txn.sign(senderSK)
		}

	}

	txid := client.sendTransaction(signedTransactions)
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
	return state.get(key, {"unit": 0}).unit

}

//TODO: return value type
//TODO: fix bugs
func GetStateBytes(state interface{}, key interface{}) {

	switch k := key.(type) {

	case string:
		key = b64.StdEncoding.EncodeToString([]byte(k))

	}

	//TODO: what about decode?
	return state.get(key, {"bytes": ""}).bytes

}