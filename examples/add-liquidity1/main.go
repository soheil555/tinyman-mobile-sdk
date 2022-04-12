package main

import (
	"fmt"
	"tinyman-mobile-sdk/v1/client"
	"tinyman-mobile-sdk/v1/pools"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/client/v2/common"
	"github.com/algorand/go-algorand-sdk/client/v2/indexer"
	"github.com/algorand/go-algorand-sdk/types"
)

func main() {

	headers := []*common.Header{
		{
			Key:   "User-Agent",
			Value: "algosdk",
		},
	}

	algodClient, err := algod.MakeClientWithHeaders("https://node.testnet.algoexplorerapi.io", "", headers)
	indexerClient, err := indexer.MakeClientWithHeaders("https://algoindexer.testnet.algoexplorerapi.io", "", headers)

	if err != nil {
		fmt.Printf("error making algodClient: %s\n", err)
		return
	}

	// account := crypto.GenerateAccount()
	// fmt.Println(account.Address.String())

	address, err := types.DecodeAddress("5SKFXC7CO2UUBB673MGYJLOTLZ7Z6PEF6WBSJSB6AFRALZ6DDEQSAZW6NM")

	if err != nil {
		fmt.Printf("error decoding address: %s\n", err)
	}

	client, err := client.MakeTinymanTestnetClient(algodClient, indexerClient, address)

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
