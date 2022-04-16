package main

import (
	"fmt"
	"tinyman-mobile-sdk/v1/client"
	"tinyman-mobile-sdk/v1/pools"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/client/v2/common"
	"github.com/algorand/go-algorand-sdk/client/v2/indexer"
	"github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/mnemonic"
	"github.com/algorand/go-algorand-sdk/types"
	"github.com/kr/pretty"
)

func main() {

	// Hardcoding account keys is not a great practice. This is for demonstration purposes only.
	// See the README & Docs for alternative signing methods.

	if err != nil {
		fmt.Printf("error import account from private key: %s\n", err)
		return
	}

	mnemonic, err := mnemonic.FromPrivateKey(userAccount.PrivateKey)

	if err != nil {
		fmt.Printf("error generating mnemonic from private key: %s\n", err)
		return
	}

	fmt.Printf("[+]user address: %s\n", userAccount.Address.String())
	fmt.Printf("[+]mnemonic: %s\n", mnemonic)

	headers := []*common.Header{
		{
			Key:   "User-Agent",
			Value: "algosdk",
		},
	}

	algodClient, err := algod.MakeClientWithHeaders("https://node.testnet.algoexplorerapi.io", "", headers)
	if err != nil {
		fmt.Printf("error making algodClient: %s\n", err)
		return
	}

	indexerClient, err := indexer.MakeClientWithHeaders("https://algoindexer.testnet.algoexplorerapi.io", "", headers)
	if err != nil {
		fmt.Printf("error making indexerClient: %s\n", err)
		return
	}

	client, err := client.MakeTinymanTestnetClient(algodClient, indexerClient, userAccount.Address)
	// By default all subsequent operations are on behalf of userAccount

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

	// Get a quote for supplying 1000.0 TinyUSDC
	quote, err := pool.FetchMintQuote(TINYUSDC.Call(1000_000_000), nil, 0.01)
	if err != nil {
		fmt.Printf("error Fetching MintQuote: %s\n", err)
		return
	}

	pretty.Println(quote)

	// Check if we are happy with the quote...
	//TODO: in python code quote.AmountsIn[ALGO] considered to be number

	fmt.Println("quote.AmountsIn[ALGO].Amount:", quote.AmountsIn[ALGO].Amount)

	if quote.AmountsIn[ALGO].Amount < 7_000_000 {

		// Prepare the mint transactions from the quote and sign them
		transactionGroup, err := pool.PrepareMintTransactionsFromQuote(quote, types.Address{})
		if err != nil {
			fmt.Printf("error preparing mint transactions from quote: %s\n", err)
			return
		}

		transactionGroup.SignWithPrivateKey(userAccount.Address, userAccount.PrivateKey)
		_, _, err = client.Submit(transactionGroup, true)
		if err != nil {
			fmt.Printf("error submit transactions: %s\n", err)
			return
		}

		// Check if any excess liquidity asset remaining after the mint
		excess, err := pool.FetchExcessAmounts(types.Address{})
		if err != nil {
			fmt.Printf("error fetching excess amounts: %s\n", err)
			return
		}

		if amount, ok := excess[pool.LiquidityAsset]; ok {

			fmt.Printf("Excess: %v\n", amount.Amount)

			if amount.Amount > 1_000_000 {
				transactionGroup, err := pool.PrepareRedeemTransactions(amount, types.Address{})
				if err != nil {
					fmt.Printf("error preparing redeem transactions: %s\n", err)
					return
				}

				transactionGroup.SignWithPrivateKey(userAccount.Address, userAccount.PrivateKey)
				_, _, err = client.Submit(transactionGroup, true)
				if err != nil {
					fmt.Printf("error submit transactions: %s\n", err)
					return
				}

			}

		}

	}

	info, err := pool.FetchPoolPosition(types.Address{})
	if err != nil {
		fmt.Printf("error fetching pool position: %s\n", err)
		return
	}
	fmt.Printf("info: %v\n", info)

	//TODO: is info["share"] float64 or what
	share := info["share"].(uint64) * 100

	fmt.Printf("Pool Tokens: %v\n", info[pool.LiquidityAsset])
	fmt.Printf("Assets: %v, %v\n", info[TINYUSDC], info[ALGO])
	fmt.Printf("share of pool: %d\n", share)

}
