package utils

import (
	"fmt"
	"testing"
	myTypes "tinyman-mobile-sdk/types"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
	"github.com/algorand/go-algorand-sdk/encoding/msgpack"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestGetProgram(t *testing.T) {

	definition := myTypes.Definition{
		Bytecode: "BCAIAQCBgICAgICAgPABgICAgICAgIDwAQMEBQYlJA1EMQkyAxJEMRUyAxJEMSAyAxJEMgQiDUQzAQAxABJEMwEQIQcSRDMBGIGCgICAgICAgPABEkQzARkiEjMBGyEEEhA3ARoAgAlib290c3RyYXASEEAAXDMBGSMSRDMBG4ECEjcBGgCABHN3YXASEEACOzMBGyISRDcBGgCABG1pbnQSQAE7NwEaAIAEYnVybhJAAZg3ARoAgAZyZWRlZW0SQAJbNwEaAIAEZmVlcxJAAnkAIQYhBSQjEk0yBBJENwEaARclEjcBGgIXJBIQRDMCADEAEkQzAhAhBBJEMwIhIxJEMwIiIxwSRDMCIyEHEkQzAiQjEkQzAiWACFRNUE9PTDExEkQzAiZRAA+AD1RpbnltYW5Qb29sMS4xIBJEMwIngBNodHRwczovL3RpbnltYW4ub3JnEkQzAikyAxJEMwIqMgMSRDMCKzIDEkQzAiwyAxJEMwMAMQASRDMDECEFEkQzAxElEkQzAxQxABJEMwMSIxJEJCMTQAAQMwEBMwIBCDMDAQg1AUIBsTMEADEAEkQzBBAhBRJEMwQRJBJEMwQUMQASRDMEEiMSRDMBATMCAQgzAwEIMwQBCDUBQgF8MgQhBhJENwEcATEAE0Q3ARwBMwQUEkQzAgAxABNEMwIUMQASRDMDADMCABJEMwIRJRJEMwMUMwMHMwMQIhJNMQASRDMDESMzAxAiEk0kEkQzBAAxABJEMwQUMwIAEkQzAQEzBAEINQFCAREyBCEGEkQ3ARwBMQATRDcBHAEzAhQSRDMDFDMDBzMDECISTTcBHAESRDMCADEAEkQzAhQzBAASRDMCESUSRDMDADEAEkQzAxQzAwczAxAiEk0zBAASRDMDESMzAxAiEk0kEkQzBAAxABNEMwQUMQASRDMBATMCAQgzAwEINQFCAJAyBCEFEkQ3ARwBMQATRDMCADcBHAESRDMCADEAE0QzAwAxABJEMwIUMwIHMwIQIhJNMQASRDMDFDMDBzMDECISTTMCABJEMwEBMwMBCDUBQgA+MgQhBBJENwEcATEAE0QzAhQzAgczAhAiEk03ARwBEkQzAQEzAgEINQFCABIyBCEEEkQzAQEzAgEINQFCAAAzAAAxABNEMwAHMQASRDMACDQBD0M=",
		Variables: []struct {
			Name   string "json:\"name\""
			Type   string "json:\"type\""
			Index  int    "json:\"index\""
			Length int    "json:\"length\""
		}{
			{
				Name:   "TMPL_ASSET_ID_1",
				Type:   "int",
				Index:  15,
				Length: 10,
			},
			{
				Name:   "TMPL_ASSET_ID_2",
				Type:   "int",
				Index:  5,
				Length: 10,
			},
			{
				Name:   "TMPL_VALIDATOR_APP_ID",
				Type:   "int",
				Index:  74,
				Length: 10,
			},
		},
	}

	variables := map[string]interface{}{
		"validator_app_id": 10,
		"asset_id_1":       2,
		"asset_id_2":       1,
	}

	expected := []byte{4, 32, 8, 1, 0, 1, 2, 3, 4, 5, 6, 37, 36, 13, 68, 49, 9, 50, 3, 18, 68, 49, 21, 50, 3, 18, 68, 49, 32, 50, 3, 18, 68, 50, 4, 34, 13, 68, 51, 1, 0, 49, 0, 18, 68, 51, 1, 16, 33, 7, 18, 68, 51, 1, 24, 129, 10, 18, 68, 51, 1, 25, 34, 18, 51, 1, 27, 33, 4, 18, 16, 55, 1, 26, 0, 128, 9, 98, 111, 111, 116, 115, 116, 114, 97, 112, 18, 16, 64, 0, 92, 51, 1, 25, 35, 18, 68, 51, 1, 27, 129, 2, 18, 55, 1, 26, 0, 128, 4, 115, 119, 97, 112, 18, 16, 64, 2, 59, 51, 1, 27, 34, 18, 68, 55, 1, 26, 0, 128, 4, 109, 105, 110, 116, 18, 64, 1, 59, 55, 1, 26, 0, 128, 4, 98, 117, 114, 110, 18, 64, 1, 152, 55, 1, 26, 0, 128, 6, 114, 101, 100, 101, 101, 109, 18, 64, 2, 91, 55, 1, 26, 0, 128, 4, 102, 101, 101, 115, 18, 64, 2, 121, 0, 33, 6, 33, 5, 36, 35, 18, 77, 50, 4, 18, 68, 55, 1, 26, 1, 23, 37, 18, 55, 1, 26, 2, 23, 36, 18, 16, 68, 51, 2, 0, 49, 0, 18, 68, 51, 2, 16, 33, 4, 18, 68, 51, 2, 33, 35, 18, 68, 51, 2, 34, 35, 28, 18, 68, 51, 2, 35, 33, 7, 18, 68, 51, 2, 36, 35, 18, 68, 51, 2, 37, 128, 8, 84, 77, 80, 79, 79, 76, 49, 49, 18, 68, 51, 2, 38, 81, 0, 15, 128, 15, 84, 105, 110, 121, 109, 97, 110, 80, 111, 111, 108, 49, 46, 49, 32, 18, 68, 51, 2, 39, 128, 19, 104, 116, 116, 112, 115, 58, 47, 47, 116, 105, 110, 121, 109, 97, 110, 46, 111, 114, 103, 18, 68, 51, 2, 41, 50, 3, 18, 68, 51, 2, 42, 50, 3, 18, 68, 51, 2, 43, 50, 3, 18, 68, 51, 2, 44, 50, 3, 18, 68, 51, 3, 0, 49, 0, 18, 68, 51, 3, 16, 33, 5, 18, 68, 51, 3, 17, 37, 18, 68, 51, 3, 20, 49, 0, 18, 68, 51, 3, 18, 35, 18, 68, 36, 35, 19, 64, 0, 16, 51, 1, 1, 51, 2, 1, 8, 51, 3, 1, 8, 53, 1, 66, 1, 177, 51, 4, 0, 49, 0, 18, 68, 51, 4, 16, 33, 5, 18, 68, 51, 4, 17, 36, 18, 68, 51, 4, 20, 49, 0, 18, 68, 51, 4, 18, 35, 18, 68, 51, 1, 1, 51, 2, 1, 8, 51, 3, 1, 8, 51, 4, 1, 8, 53, 1, 66, 1, 124, 50, 4, 33, 6, 18, 68, 55, 1, 28, 1, 49, 0, 19, 68, 55, 1, 28, 1, 51, 4, 20, 18, 68, 51, 2, 0, 49, 0, 19, 68, 51, 2, 20, 49, 0, 18, 68, 51, 3, 0, 51, 2, 0, 18, 68, 51, 2, 17, 37, 18, 68, 51, 3, 20, 51, 3, 7, 51, 3, 16, 34, 18, 77, 49, 0, 18, 68, 51, 3, 17, 35, 51, 3, 16, 34, 18, 77, 36, 18, 68, 51, 4, 0, 49, 0, 18, 68, 51, 4, 20, 51, 2, 0, 18, 68, 51, 1, 1, 51, 4, 1, 8, 53, 1, 66, 1, 17, 50, 4, 33, 6, 18, 68, 55, 1, 28, 1, 49, 0, 19, 68, 55, 1, 28, 1, 51, 2, 20, 18, 68, 51, 3, 20, 51, 3, 7, 51, 3, 16, 34, 18, 77, 55, 1, 28, 1, 18, 68, 51, 2, 0, 49, 0, 18, 68, 51, 2, 20, 51, 4, 0, 18, 68, 51, 2, 17, 37, 18, 68, 51, 3, 0, 49, 0, 18, 68, 51, 3, 20, 51, 3, 7, 51, 3, 16, 34, 18, 77, 51, 4, 0, 18, 68, 51, 3, 17, 35, 51, 3, 16, 34, 18, 77, 36, 18, 68, 51, 4, 0, 49, 0, 19, 68, 51, 4, 20, 49, 0, 18, 68, 51, 1, 1, 51, 2, 1, 8, 51, 3, 1, 8, 53, 1, 66, 0, 144, 50, 4, 33, 5, 18, 68, 55, 1, 28, 1, 49, 0, 19, 68, 51, 2, 0, 55, 1, 28, 1, 18, 68, 51, 2, 0, 49, 0, 19, 68, 51, 3, 0, 49, 0, 18, 68, 51, 2, 20, 51, 2, 7, 51, 2, 16, 34, 18, 77, 49, 0, 18, 68, 51, 3, 20, 51, 3, 7, 51, 3, 16, 34, 18, 77, 51, 2, 0, 18, 68, 51, 1, 1, 51, 3, 1, 8, 53, 1, 66, 0, 62, 50, 4, 33, 4, 18, 68, 55, 1, 28, 1, 49, 0, 19, 68, 51, 2, 20, 51, 2, 7, 51, 2, 16, 34, 18, 77, 55, 1, 28, 1, 18, 68, 51, 1, 1, 51, 2, 1, 8, 53, 1, 66, 0, 18, 50, 4, 33, 4, 18, 68, 51, 1, 1, 51, 2, 1, 8, 53, 1, 66, 0, 0, 51, 0, 0, 49, 0, 19, 68, 51, 0, 7, 49, 0, 18, 68, 51, 0, 8, 52, 1, 15, 67}

	result, err := GetProgram(definition, variables)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, expected)
	assert.Equal(t, expected, result)

}

func TestEncodeVarint(t *testing.T) {

	var input uint64 = 123123
	expected := []byte{243, 193, 7}

	assert.Equal(t, expected, EncodeVarint(input))

}

func TestEncodeValue(t *testing.T) {

	var input interface{} = 111111
	wrongInput := "wrong"
	expected := []byte{135, 228, 6}

	_, err := EncodeValue(input, "string")
	assert.NotNil(t, err)

	_, err = EncodeValue(wrongInput, "int")
	assert.NotNil(t, err)

	result, err := EncodeValue(input, "int")
	assert.Nil(t, err)

	assert.Equal(t, expected, result)

}

func TestIntToBytes(t *testing.T) {

	var input uint64 = 123123123123123
	expected := []byte{0, 0, 111, 250, 214, 4, 115, 179}

	assert.Equal(t, expected, IntToBytes(input))

}

func TestGetStateInt(t *testing.T) {

	state := map[string]models.TealValue{
		"YTE=": {
			Uint: 1,
		},
		"YTI=": {
			Uint: 2,
		},
	}

	assert.Equal(t, uint64(1), GetStateInt(state, "a1"))
	assert.Equal(t, uint64(2), GetStateInt(state, "a2"))
	assert.Equal(t, uint64(2), GetStateInt(state, []byte{89, 84, 73, 61}))
	assert.Equal(t, uint64(0), GetStateInt(state, "a3"))

}

func TestGetStateBytes(t *testing.T) {

	state := map[string]models.TealValue{
		"YTE=": {
			Bytes: "test1",
		},
		"YTI=": {
			Bytes: "test2",
		},
	}

	assert.Equal(t, "test1", GetStateBytes(state, "a1"))
	assert.Equal(t, "test2", GetStateBytes(state, "a2"))
	assert.Equal(t, "test2", GetStateBytes(state, []byte{89, 84, 73, 61}))
	assert.Equal(t, "", GetStateBytes(state, "a3"))

}

// func TestSignAndSubmitTransactions(t *testing.T) {

// }

func TestWaitForConfirmation(t *testing.T) {

	defer gock.Off()

	mockServerURL := "https://mockserver.com"
	lastRound := uint64(1)
	txid := "4"

	gock.New(mockServerURL).Get("/v2/status").
		Reply(200).
		JSON(map[string]uint64{
			"last-round": lastRound,
		})

	gock.New(mockServerURL).Get(fmt.Sprintf("/v2/status/wait-for-block-after/%v", lastRound+1)).
		Reply(200).JSON(map[string]uint64{
		"last-round": lastRound + 1,
	})

	gock.New(mockServerURL).Get(fmt.Sprintf("/v2/transactions/pending/%s", txid)).
		Reply(200).JSON(msgpack.Encode(map[string]uint64{
		"confirmed-round": 0,
	}))

	gock.New(mockServerURL).Get(fmt.Sprintf("/v2/transactions/pending/%s", txid)).
		Reply(200).JSON(msgpack.Encode(map[string]uint64{
		"confirmed-round": lastRound + 1,
	}))

	mockClient, err := algod.MakeClient(mockServerURL, "")
	assert.Nil(t, err)

	result, resultTxid, err := WaitForConfirmation(mockClient, txid)
	assert.Nil(t, err)
	assert.Equal(t, txid, resultTxid)
	assert.Equal(t, lastRound+1, result.ConfirmedRound)

}
