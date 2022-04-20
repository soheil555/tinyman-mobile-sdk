package main

import (
	"context"
	"fmt"
	"os"
	"tinyman-mobile-sdk/types"
	"tinyman-mobile-sdk/utils"
	"tinyman-mobile-sdk/v1/client"
	"tinyman-mobile-sdk/v1/pools"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/client/v2/common"
	"github.com/algorand/go-algorand-sdk/client/v2/indexer"
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

	//TODO: in python code validatorID is for v1.0
	client := client.MakeTinymanClient(algodClient, indexerClient, 62368684, algoTypes.Address{})
	// By default all subsequent operations are on behalf of userAccount

	// Check if the account is opted into Tinyman and optin if necessary
	isOptedIn, err := client.IsOptedIn(userAccount.Address)

	if err != nil {
		fmt.Printf("error checking if the user has opted into Tinyman: %s\n", err)
		return
	}

	if !isOptedIn {

		fmt.Println("Account not opted into app, opting in now...")

		transactionGroup, err := client.PrepareAppOptinTransactions(userAccount.Address)
		if err != nil {
			fmt.Printf("error preparing app optin transactions: %s\n", err)
			return
		}

		for i, txn := range transactionGroup.Transactions {

			if txn.Sender == userAccount.Address {

				_, stxBytes, err := crypto.SignTransaction(privateKey, txn)
				if err != nil {
					fmt.Printf("failed to sign transaction: %s\n", err)
					return
				}
				transactionGroup.SignedTransactions[i] = stxBytes

			}

		}

		var signedGroup []byte

		for _, txn := range transactionGroup.SignedTransactions {
			signedGroup = append(signedGroup, txn...)
		}

		sendRawTransaction := client.Algod.SendRawTransaction(signedGroup)
		txid, err := sendRawTransaction.Do(context.Background())
		if err != nil {
			fmt.Printf("error submitting optin transactionGroup: %s\n", err)
			return
		}

		_, _, err = utils.WaitForConfirmation(client.Algod, txid)
		if err != nil {
			fmt.Printf("error waiting for optin transactionGroup confirmation: %s\n", err)
			return
		}

	}

	// Fetch our two assets of interest
	TINYUSDC := types.Asset{Id: 21582668, Name: "TinyUSDC", UnitName: "INYUSDC", Decimals: 6}
	ALGO := types.Asset{Id: 0, Name: "Algo", UnitName: "ALGO", Decimals: 6}

	// Create the pool we will work with and fetch its on-chain state
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

	priceWithSlippage, err := quote.PriceWithSlippage()
	if err != nil {
		fmt.Printf("error fetching price with slippage: %s\n", err)
		return
	}

	// We only want to sell if ALGO is > 180 TINYUSDC (It's testnet!)
	if priceWithSlippage > 110 {
		fmt.Printf("Swapping %v to %v\n", quote.AmountIn, priceWithSlippage)

		// Prepare a transaction group
		amountOutWithSlippage, err := quote.AmountOutWithSlippage()
		if err != nil {
			fmt.Printf("error fetching amount out with slippage: %s\n", err)
			return
		}

		transactionGroup, err := pool.PrepareSwapTransactions(quote.AmountIn, amountOutWithSlippage, "fixed-input", userAccount.Address)
		if err != nil {
			fmt.Printf("error preparing swap transactions: %s\n", err)
			return
		}

		for i, txn := range transactionGroup.Transactions {

			if txn.Sender == userAccount.Address {

				_, stxBytes, err := crypto.SignTransaction(privateKey, txn)
				if err != nil {
					fmt.Printf("failed to sign transaction: %s\n", err)
					return
				}
				transactionGroup.SignedTransactions[i] = stxBytes

			}

		}

		var signedGroup []byte

		for _, txn := range transactionGroup.SignedTransactions {
			signedGroup = append(signedGroup, txn...)
		}

		sendRawTransaction := client.Algod.SendRawTransaction(signedGroup)
		txid, err := sendRawTransaction.Do(context.Background())
		if err != nil {
			fmt.Printf("error submitting swap transactionGroup: %s\n", err)
			return
		}

		_, _, err = utils.WaitForConfirmation(client.Algod, txid)
		if err != nil {
			fmt.Printf("error waiting for swap transactionGroup confirmation: %s\n", err)
			return
		}

		// Check if any excess remaining after the swap
		excessAmounts, err := pool.FetchExcessAmounts(userAccount.Address)
		if err != nil {
			fmt.Printf("error fetching excess amounts: %v\n", excessAmounts)
			return
		}

		if excess, ok := excessAmounts[TINYUSDC]; ok {

			fmt.Printf("Excess: %d\n", excess.Amount)

			// We might just let the excess accumulate rather than redeeming if its < 1 TinyUSDC
			if excess.Amount > 1_000 {

				fmt.Println("redeeming excess amount...")
				transactionGroup, err := pool.PrepareRedeemTransactions(excess, userAccount.Address)
				if err != nil {
					fmt.Printf("error preparing redeem transactions: %s\n", err)
					return
				}

				// Sign the group with our key
				for i, txn := range transactionGroup.Transactions {

					if txn.Sender == userAccount.Address {

						_, stxBytes, err := crypto.SignTransaction(privateKey, txn)
						if err != nil {
							fmt.Printf("failed to sign transaction: %s\n", err)
							return
						}
						transactionGroup.SignedTransactions[i] = stxBytes

					}

				}

				var signedGroup []byte

				for _, txn := range transactionGroup.SignedTransactions {
					signedGroup = append(signedGroup, txn...)
				}

				sendRawTransaction := client.Algod.SendRawTransaction(signedGroup)
				txid, err := sendRawTransaction.Do(context.Background())
				if err != nil {
					fmt.Printf("error submitting redeem transactionGroup: %s\n", err)
					return
				}

				_, _, err = utils.WaitForConfirmation(client.Algod, txid)
				if err != nil {
					fmt.Printf("error waiting for redeem transactionGroup confirmation: %s\n", err)
					return
				}

			}

		}

	}

}
