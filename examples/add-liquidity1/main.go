package main

import (
	"fmt"
	"tinyman-mobile-sdk/v1/client"
	"tinyman-mobile-sdk/v1/pools"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/client/v2/common"
	"github.com/algorand/go-algorand-sdk/crypto"
)

func main() {

	headers := []*common.Header{
		{
			Key:   "User-Agent",
			Value: "algosdk",
		},
	}

	algodClient, err := algod.MakeClientWithHeaders("https://node.algoexplorerapi.io/", "", headers)

	if err != nil {
		fmt.Printf("error making algodClient: %s\n", err)
		return
	}

	account := crypto.GenerateAccount()

	client, err := client.MakeTinymanTestnetClient(algodClient, account.Address)

	if err != nil {
		fmt.Printf("error making tinyManTestnetClient: %s\n", err)
		return
	}

	// Fetch our two assets of interest
	TINYUSDC := client.FetchAsset(21582668)
	ALGO := client.FetchAsset(0)

	// Fetch the pool we will work with
	//TODO: make pool from client
	pool, err := pools.MakePool(client, TINYUSDC, ALGO, nil, true, nil)

	if err != nil {
		fmt.Printf("error making pool: %s\n", err)
		return
	}

	quote, err := pool.FetchMintQuote(TINYUSDC.Call(1000_000_000), nil, 0.01)
	if err != nil {
		fmt.Printf("error Fetching MintQuote: %s\n", err)
		return
	}
	fmt.Println(quote)

}
