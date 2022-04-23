package main

import (
	"fmt"
	"os"

	"github.com/soheil555/tinyman-mobile-sdk/v1/client"
	"github.com/soheil555/tinyman-mobile-sdk/v1/pools"

	"github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/mnemonic"
	algoTypes "github.com/algorand/go-algorand-sdk/types"
	"github.com/joho/godotenv"
)

// This sample is provided for demonstration purposes only.
// It is not intended for production use.
// This example does not constitute trading advice.

func main() {

	err := godotenv.Load("../../.env")

	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}

	// Hardcoding account keys is not a great practice. This is for demonstration purposes only.
	// See the README & Docs for alternative signing methods.

	privateKey, err := mnemonic.ToPrivateKey(os.Getenv("MNEMONIC"))

	if err != nil {
		fmt.Printf("error generating private key from mnemonic: %s\n", err)
		return
	}

	userAccount, err := crypto.AccountFromPrivateKey(privateKey)

	if err != nil {
		fmt.Printf("error import account from private key: %s\n", err)
		return
	}

	fmt.Printf("[+]user address: %s\n", userAccount.Address.String())

	algodClientURL := "https://node.testnet.algoexplorerapi.io"
	indexerClientURL := "https://algoindexer.testnet.algoexplorerapi.io"

	client, err := client.NewTinymanTestnetClient(algodClientURL, indexerClientURL, userAccount.Address)
	// By default all subsequent operations are on behalf of userAccount

	if err != nil {
		fmt.Printf("error making tinyManTestnetClient: %s\n", err)
		return
	}

	// Fetch our two assets of interest
	TINYUSDC, err := client.FetchAsset(21582668)
	if err != nil {
		fmt.Printf("error fetching asset: %s\n", err)
		return
	}

	ALGO, err := client.FetchAsset(0)
	if err != nil {
		fmt.Printf("error fetching asset: %s\n", err)
		return
	}

	// Fetch the pool we will work with
	//TODO: make pool from client
	pool, err := pools.NewPool(client, TINYUSDC, ALGO, pools.PoolInfo{}, true, 0)

	if err != nil {
		fmt.Printf("error making pool: %s\n", err)
		return
	}

	info, share, err := pool.FetchPoolPosition(algoTypes.Address{})
	if err != nil {
		fmt.Printf("error fetching pool position: %s\n", err)
		return
	}

	fmt.Printf("Pool Tokens: %v\n", info[pool.LiquidityAsset])
	fmt.Printf("Assets: %v, %v\n", info[TINYUSDC], info[ALGO])
	fmt.Printf("share of pool: %.3f\n", share*100)

}
