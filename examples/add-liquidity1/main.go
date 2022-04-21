package main

import (
	"fmt"
	"os"
	"tinyman-mobile-sdk/types"
	"tinyman-mobile-sdk/v1/client"
	"tinyman-mobile-sdk/v1/pools"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/client/v2/common"
	"github.com/algorand/go-algorand-sdk/client/v2/indexer"
	"github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/mnemonic"
	algoTypes "github.com/algorand/go-algorand-sdk/types"
	"github.com/joho/godotenv"
	"github.com/kr/pretty"
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

	client, err := client.NewTinymanTestnetClient(algodClient, indexerClient, userAccount.Address)
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

	// Get a quote for supplying 10.0 TinyUSDC
	quote, err := pool.FetchMintQuote(TINYUSDC.Call(10_000_000), types.AssetAmount{}, 0.01)
	if err != nil {
		fmt.Printf("error Fetching MintQuote: %s\n", err)
		return
	}

	//TODO: in some places we use pointer but in some places, we don't. what to do
	pretty.Println(quote)

	// Check if we are happy with the quote...
	//TODO: in python code quote.AmountsIn[ALGO] considered to be number
	if quote.AmountsIn[ALGO].Amount < 1_000_000 {

		// Prepare the mint transactions from the quote and sign them
		transactionGroup, err := pool.PrepareMintTransactionsFromQuote(quote, algoTypes.Address{})
		if err != nil {
			fmt.Printf("error preparing mint transactions from quote: %s\n", err)
			return
		}

		transactionGroup.SignWithPrivateKey(userAccount.Address, userAccount.PrivateKey)
		_, _, err = client.Submit(transactionGroup, true)
		if err != nil {
			fmt.Printf("error submit transactions 1: %s\n", err)
			return
		}

		// Check if any excess liquidity asset remaining after the mint
		excess, err := pool.FetchExcessAmounts(algoTypes.Address{})
		if err != nil {
			fmt.Printf("error fetching excess amounts: %s\n", err)
			return
		}

		if amount, ok := excess[pool.LiquidityAsset]; ok {

			fmt.Printf("Excess: %v\n", amount.Amount)

			if amount.Amount > 1_000 {

				transactionGroup, err := pool.PrepareRedeemTransactions(amount, algoTypes.Address{})
				if err != nil {
					fmt.Printf("error preparing redeem transactions: %s\n", err)
					return
				}

				transactionGroup.SignWithPrivateKey(userAccount.Address, userAccount.PrivateKey)
				_, _, err = client.Submit(transactionGroup, true)
				if err != nil {
					fmt.Printf("error submit transactions 2: %s\n", err)
					return
				}

			}

		}

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
