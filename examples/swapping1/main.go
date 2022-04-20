package main

import (
	"fmt"
	"os"
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

	client, err := client.MakeTinymanTestnetClient(algodClient, indexerClient, userAccount.Address)
	// By default all subsequent operations are on behalf of userAccount

	if err != nil {
		fmt.Printf("error making tinyManTestnetClient: %s\n", err)
		return
	}

	// Check if the account is opted into Tinyman and optin if necessary
	isOptedIn, err := client.IsOptedIn(algoTypes.Address{})

	if err != nil {
		fmt.Printf("error checking if the user has opted into Tinyman: %s\n", err)
		return
	}

	if !isOptedIn {

		fmt.Println("Account not opted into app, opting in now...")

		transactionGroup, err := client.PrepareAppOptinTransactions(algoTypes.Address{})
		if err != nil {
			fmt.Printf("error preparing app optin transactions: %s\n", err)
			return
		}

		err = transactionGroup.SignWithPrivateKey(userAccount.Address, userAccount.PrivateKey)
		if err != nil {
			fmt.Printf("error signing optin transactionGroup: %s\n", err)
			return
		}

		_, _, err = client.Submit(transactionGroup, true)
		if err != nil {
			fmt.Printf("error submitting optin transactionGroup: %s\n", err)
			return
		}

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
	pool, err := pools.MakePool(client, TINYUSDC, ALGO, pools.PoolInfo{}, true, 0)
	if err != nil {
		fmt.Printf("error making pool: %s\n", err)
		return
	}

	// Get a quote for a swap of 1 ALGO to TINYUSDC with 1% slippage tolerance
	quote, err := pool.FetchFixedInputSwapQuote(ALGO.Call(1_000_000), 0.01)
	if err != nil {
		fmt.Printf("error fetching fixed input swap quote: %s\n", err)
		return
	}

	pretty.Println(quote)
	fmt.Printf("TINYUSDC per ALGO: %f\n", quote.Price())
	priceWithSlippage, err := quote.PriceWithSlippage()

	if err != nil {
		fmt.Printf("error getting price with slippage: %s\n", err)
	}

	fmt.Printf("TINYUSDC per ALGO (worst case): %f\n", priceWithSlippage)

	// We only want to sell if ALGO is > 180 TINYUSDC (It's testnet!)
	if priceWithSlippage > 110 {

		amountOutWithSlippage, err := quote.AmountOutWithSlippage()
		if err != nil {
			fmt.Printf("error getting amout out with slippage: %v\n", amountOutWithSlippage)
			return
		}

		fmt.Printf("Swapping %v to %v\n", quote.AmountIn, amountOutWithSlippage)

		// Prepare a transaction group
		transactionGroup, err := pool.PrepareSwapTransactionsFromQuote(quote, algoTypes.Address{})
		if err != nil {
			fmt.Printf("error preparing swap transactions from quote: %s\n", err)
			return
		}

		// Sign the group with our key
		err = transactionGroup.SignWithPrivateKey(userAccount.Address, privateKey)
		if err != nil {
			fmt.Printf("error signing swap transactionGroup: %s\n", err)
			return
		}

		// Submit transactions to the network and wait for confirmation
		_, _, err = client.Submit(transactionGroup, true)
		if err != nil {
			fmt.Printf("error submitting swap transactionGroup: %s\n", err)
			return
		}

		// Check if any excess remaining after the swap
		excessAmounts, err := pool.FetchExcessAmounts(algoTypes.Address{})
		if err != nil {
			fmt.Printf("error fetching excess amounts: %v\n", excessAmounts)
			return
		}

		if excess, ok := excessAmounts[TINYUSDC]; ok {

			fmt.Printf("Excess: %d \n", excess.Amount)

			// We might just let the excess accumulate rather than redeeming if its < 1 TinyUSDC
			if excess.Amount > 1_000 {

				fmt.Println("redeeming excess amount...")

				transactionGroup, err := pool.PrepareRedeemTransactions(excess, algoTypes.Address{})
				if err != nil {
					fmt.Printf("error preparing redeem transactions: %s\n", err)
					return
				}

				err = transactionGroup.SignWithPrivateKey(userAccount.Address, userAccount.PrivateKey)
				if err != nil {
					fmt.Printf("error signing redeem transactionGroup with private key: %s\n", err)
					return
				}

				_, _, err = client.Submit(transactionGroup, true)
				if err != nil {
					fmt.Printf("error submitting reddem transactionGroup: %s\n", err)
					return
				}

			}

		}

	}

}
