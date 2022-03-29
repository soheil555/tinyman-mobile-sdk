package utils

import (
	"context"
	"crypto/ed25519"
	b64 "encoding/base64"
	"encoding/binary"
	"fmt"
	"sort"
	"strings"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/logic"
	"github.com/algorand/go-algorand-sdk/transaction"
	"github.com/algorand/go-algorand-sdk/types"
)

type Variable struct {
	name   string
	_type  string
	index  int
	length int
}

type Definition struct {
	bytecode  string
	address   string
	size      int
	variables []Variable
	source    string
}

/*
	Return a byte array to be used in LogicSig.
*/
func GetProgram(definition Definition, variables map[string]interface{}) ([]byte, error) {

	template := definition.bytecode

	templateBytes, err := b64.StdEncoding.DecodeString(template)
	if err != nil {
		return nil, err
	}

	offset := 0
	var dVariables = definition.variables

	sort.SliceStable(dVariables, func(i, j int) bool {
		return dVariables[i].index < dVariables[j].index
	})

	for _, v := range dVariables {

		s := strings.Split(v.name, "TMPL_")
		name := strings.ToLower(s[len(s)-1])
		value := variables[name]
		start := v.index - offset
		end := start + v.length
		valueEncoded, err := EncodeValue(value, v._type)

		if err != nil {
			return nil, err
		}

		valueEncodedLen := len(valueEncoded)
		diff := v.length - valueEncodedLen
		offset += diff
		//TODO: better way for assign?
		for i := start; i < end; i++ {
			templateBytes[i] = valueEncoded[i-start]
		}

	}

	return templateBytes, nil

}

func EncodeValue(value interface{}, _type string) ([]byte, error) {

	if _type == "int" {
		return EncodeVarint(value.(int)), nil
	} else {
		return nil, fmt.Errorf("Unsupported value type %s!", _type)
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

func SignAndSubmitTransactions(client algod.Client, transactions []types.Transaction, signedTransactions [][]byte, sender types.Address, senderSK ed25519.PrivateKey) (*algod.PendingTransactionInformation, error) {

	for i, txn := range transactions {

		if txn.Sender == sender {
			_, stx, err := crypto.SignTransaction(senderSK, txn)

			if err != nil {
				return nil, fmt.Errorf("Signing failed with %v", err)
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
		return nil, fmt.Errorf("Failed to create transaction: %v\n", err)
	}

	return WaitForConfirmation(client, txid)

}

/*
   Utility function to wait until the transaction is
   confirmed before proceeding.
*/
func WaitForConfirmation(client algod.Client, txid string) (*algod.PendingTransactionInformation, error) {

	nodeStatus, err := client.Status().Do(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error getting algod status: %s\n", err)
	}

	lastRound := nodeStatus.LastRound

	txinfo := client.PendingTransactionInformation(txid)

	pendingTrxInfo, _, err := txinfo.Do(context.Background())

	if err != nil {
		return nil, fmt.Errorf("error getting algod pending transaction info: %s\n", err)
	}

	for !(pendingTrxInfo.ConfirmedRound > 0) {

		fmt.Println("Waiting for confirmation")
		lastRound += 1
		client.StatusAfterBlock(lastRound)

		pendingTrxInfo, _, err = txinfo.Do(context.Background())

		if err != nil {
			return nil, fmt.Errorf("error getting algod pending transaction info: %s\n", err)
		}

	}

	//TODO: what should return and how to set txid
	// txinfo.txid = txid

	fmt.Printf("Transaction %s confirmed in round %d.\n", txid, pendingTrxInfo.ConfirmedRound)
	return txinfo, nil

}

func IntToBytes(num uint) []byte {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data[:], uint64(num))
	return data
}

//TODO: fix bugs
func GetStateInt(state interface{}, key interface{}) {

	switch k := key.(type) {

	case string:
		key = b64.StdEncoding.EncodeToString([]byte(k))

	}

}

//TODO: fix bugs
func GetStateBytes(state interface{}, key interface{}) {

	switch k := key.(type) {

	case string:
		key = b64.StdEncoding.EncodeToString([]byte(k))

	}

}

//TODO: should move to another file?
type transactionGroup struct {
	transactions       []types.Transaction
	signedTransactions [][]byte
}

func NewTransactionGroup(transactions []types.Transaction) (transactionGroup, error) {

	transactions, err := transaction.AssignGroupID(transactions, "")
	if err != nil {
		return transactionGroup{}, err
	}
	//TODO: [][]byte. is it good?
	signedTransactions := make([][]byte, len(transactions))
	return transactionGroup{transactions, signedTransactions}, nil

}

//TODO: what is user?
// func (s *transactionGroup) Sign(user interface{}) {
// 	user.signTransactionGroup(s)
// }

func (s *transactionGroup) SignWithLogicsig(logicsig types.LogicSig) error {

	_, byteArrays, err := logic.ReadProgram(logicsig.Logic, nil)

	if err != nil {
		return err
	}

	//TODO: where is address in byteArray?
	var address types.Address

	n := copy(address[:], byteArrays[1])

	if n != ed25519.PublicKeySize {
		return fmt.Errorf("address generated from receiver bytes is the wrong size")
	}

	for i, txn := range s.transactions {
		if txn.Sender == address {
			_, stxBytes, err := crypto.SignLogicsigTransaction(logicsig, txn)

			if err != nil {
				return fmt.Errorf("Failed to create transaction: %v\n", err)
			}

			s.signedTransactions[i] = stxBytes
		}
	}

	return nil

}

func (s *transactionGroup) SignWithPrivateKey(address types.Address, privateKey ed25519.PrivateKey) error {

	for i, txn := range s.transactions {
		if txn.Sender == address {
			_, stxBytes, err := crypto.SignTransaction(privateKey, txn)
			if err != nil {
				return fmt.Errorf("Failed to sign transaction: %v\n", err)
			}
			s.signedTransactions[i] = stxBytes
		}
	}

	return nil

}

func (s *transactionGroup) Sumbit(algod algod.Client, wait bool) (*algod.PendingTransactionInformation, error) {

	var signedGroup []byte

	for _, txn := range s.signedTransactions {

		signedGroup = append(signedGroup, txn...)

	}

	txid, err := algod.SendRawTransaction(signedGroup).Do(context.Background())

	if err != nil {
		return nil, fmt.Errorf("Failed to send transaction: %v\n", err)
	}

	if wait {
		return WaitForConfirmation(algod, txid)
	}

	//TODO: we need to return txid as a struct
	return nil, nil

}
