package utils

import (
	"context"
	"crypto/ed25519"
	b64 "encoding/base64"
	"encoding/binary"
	"fmt"
	"sort"
	"strings"
	myTypes "tinyman-mobile-sdk/types"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
	"github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/logic"
	"github.com/algorand/go-algorand-sdk/transaction"
	"github.com/algorand/go-algorand-sdk/types"
)

/*
	Return a byte array to be used in LogicSig.
*/
func GetProgram(definition myTypes.Definition, variables map[string]interface{}) ([]byte, error) {

	template := definition.Bytecode

	templateBytes, err := b64.StdEncoding.DecodeString(template)
	if err != nil {
		return nil, err
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
		valueEncoded, err := EncodeValue(value, v.Type)

		if err != nil {
			return nil, err
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

	return templateBytes, nil

}

//TODO: is value interface{} ok?
//TODO: what about type assertion
func EncodeValue(value interface{}, _type string) ([]byte, error) {

	if _type == "int" {
		value, ok := value.(int)
		if !ok {
			return nil, fmt.Errorf("type assertion value to int failed")
		}
		return EncodeVarint(uint64(value)), nil
	} else {
		return nil, fmt.Errorf("unsupported value type %s", _type)
	}

}

func EncodeVarint(number uint64) []byte {

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

func SignAndSubmitTransactions(client *algod.Client, transactions []types.Transaction, signedTransactions [][]byte, sender types.Address, senderSK ed25519.PrivateKey) (*models.PendingTransactionInfoResponse,string, error) {

	for i, txn := range transactions {

		if txn.Sender == sender {
			_, stx, err := crypto.SignTransaction(senderSK, txn)

			if err != nil {
				return nil, "",fmt.Errorf("signing failed with %v", err)
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
		return nil, "",fmt.Errorf("failed to create transaction: %v", err)
	}

	return WaitForConfirmation(client, txid)

}

/*
   Utility function to wait until the transaction is
   confirmed before proceeding.
*/
func WaitForConfirmation(client *algod.Client, txid string) (*models.PendingTransactionInfoResponse,string, error) {

	nodeStatus, err := client.Status().Do(context.Background())
	if err != nil {
		return nil, "",fmt.Errorf("error getting algod status: %s", err)
	}

	lastRound := nodeStatus.LastRound

	txinfo := client.PendingTransactionInformation(txid)

	pendingTrxInfo, _, err := txinfo.Do(context.Background())

	if err != nil {
		return nil, "",fmt.Errorf("error getting algod pending transaction info: %s", err)
	}


	//TODO: is it correct?
	for !(pendingTrxInfo.ConfirmedRound > 0) {

		fmt.Println("Waiting for confirmation")
		lastRound += 1
		//TODO: do nothing with response?
		_ , err := client.StatusAfterBlock(lastRound).Do(context.Background())
		if err != nil {
			return nil,"", fmt.Errorf("error getting algod pending transaction info: %s", err)
		}

		pendingTrxInfo, _, err = txinfo.Do(context.Background())

		if err != nil {
			return nil,"", fmt.Errorf("error getting algod pending transaction info: %s", err)
		}

	}

	//TODO: is it good way to return txid
	

	fmt.Printf("Transaction %s confirmed in round %d.\n", txid, pendingTrxInfo.ConfirmedRound)
	return &pendingTrxInfo,txid, nil

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
	Transactions       []types.Transaction
	SignedTransactions [][]byte
}

func NewTransactionGroup(transactions []types.Transaction) (TransactionGroup, error) {

	transactions, err := transaction.AssignGroupID(transactions, "")
	if err != nil {
		return TransactionGroup{}, err
	}
	//TODO: [][]byte. is it good?
	signedTransactions := make([][]byte, len(transactions))
	return TransactionGroup{transactions, signedTransactions}, nil

}

//TODO: what is user?
// func (s *transactionGroup) Sign(user interface{}) {
// 	user.signTransactionGroup(s)
// }

func (s *TransactionGroup) SignWithLogicsig(logicsig types.LogicSig) error {

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

	for i, txn := range s.Transactions {
		if txn.Sender == address {
			_, stxBytes, err := crypto.SignLogicsigTransaction(logicsig, txn)

			if err != nil {
				return fmt.Errorf("failed to create transaction: %v", err)
			}

			s.SignedTransactions[i] = stxBytes
		}
	}

	return nil

}

func (s *TransactionGroup) SignWithPrivateKey(address types.Address, privateKey ed25519.PrivateKey) error {

	for i, txn := range s.Transactions {
		if txn.Sender == address {
			_, stxBytes, err := crypto.SignTransaction(privateKey, txn)
			if err != nil {
				return fmt.Errorf("failed to sign transaction: %v", err)
			}
			s.SignedTransactions[i] = stxBytes
		}
	}

	return nil

}

func (s *TransactionGroup) Sumbit(algod *algod.Client, wait bool) (*models.PendingTransactionInfoResponse,string, error) {

	var signedGroup []byte

	for _, txn := range s.SignedTransactions {

		signedGroup = append(signedGroup, txn...)

	}

	txid, err := algod.SendRawTransaction(signedGroup).Do(context.Background())

	if err != nil {
		return nil, "",fmt.Errorf("failed to send transaction: %v", err)
	}

	if wait {
		return WaitForConfirmation(algod, txid)
	}

	//TODO: we need to return txid as a struct
	return nil, txid,nil

}
