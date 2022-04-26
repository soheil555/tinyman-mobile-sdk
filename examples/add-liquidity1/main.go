package main

import (
	"fmt"
	"math/big"
	"os"
	"strconv"

	"github.com/soheil555/tinyman-mobile-sdk/types"
	"github.com/soheil555/tinyman-mobile-sdk/v1/client"
	"github.com/soheil555/tinyman-mobile-sdk/v1/pools"

	"github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/mnemonic"
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

	algodClientURL := "https://node.testnet.algoexplorerapi.io"
	indexerClientURL := "https://algoindexer.testnet.algoexplorerapi.io"

	client, err := client.NewTinymanTestnetClient(algodClientURL, indexerClientURL, userAccount.Address.String())
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
	pool, err := pools.NewPool(client, TINYUSDC, ALGO, nil, true, 0)

	if err != nil {
		fmt.Printf("error making pool: %s\n", err)
		return
	}

	// Get a quote for supplying 10.0 TinyUSDC
	quote, err := pool.FetchMintQuote(TINYUSDC.Call("10000000"), nil, 0.01)
	if err != nil {
		fmt.Printf("error Fetching MintQuote: %s\n", err)
		return
	}

	pretty.Println(quote)

	// Check if we are happy with the quote...
	amountsIn := quote.GetAmountsIn()

	algoAmountsIn, _ := new(big.Int).SetString(amountsIn[ALGO.Id], 10)

	if algoAmountsIn.Cmp(big.NewInt(1_000_000)) < 0 {

		// Prepare the mint transactions from the quote and sign them
		transactionGroup, err := pool.PrepareMintTransactionsFromQuote(quote, "")
		if err != nil {
			fmt.Printf("error preparing mint transactions from quote: %s\n", err)
			return
		}

		transactionGroup.SignWithPrivateKey(userAccount.Address.String(), string(userAccount.PrivateKey))
		_, err = client.Submit(transactionGroup, true)
		if err != nil {
			fmt.Printf("error submit transactions 1: %s\n", err)
			return
		}

		// Check if any excess liquidity asset remaining after the mint
		excess, err := pool.FetchExcessAmounts("")
		if err != nil {
			fmt.Printf("error fetching excess amounts: %s\n", err)
			return
		}

		if amount, ok := excess[pool.LiquidityAsset.Id]; ok {

			fmt.Printf("Excess: %v\n", amount)
			amountUint, _ := new(big.Int).SetString(amount, 10)

			if amountUint.Cmp(big.NewInt(1_000)) > 0 {

				assetAmount := &types.AssetAmount{
					Asset:  pool.LiquidityAsset,
					Amount: amount,
				}

				transactionGroup, err := pool.PrepareRedeemTransactions(assetAmount, "")
				if err != nil {
					fmt.Printf("error preparing redeem transactions: %s\n", err)
					return
				}

				transactionGroup.SignWithPrivateKey(userAccount.Address.String(), string(userAccount.PrivateKey))
				_, err = client.Submit(transactionGroup, true)
				if err != nil {
					fmt.Printf("error submit transactions 2: %s\n", err)
					return
				}

			}

		}

	}

	info, err := pool.FetchPoolPosition("")
	if err != nil {
		fmt.Printf("error fetching pool position: %s\n", err)
		return
	}

	share, _ := new(big.Float).SetString(info["share"])
	shareFloat, _ := share.Float64()

	fmt.Printf("Pool Tokens: %v\n", info[strconv.Itoa(pool.LiquidityAsset.Id)])
	fmt.Printf("Assets: TINYUSDC:%v, ALGO:%v\n", info[strconv.Itoa(TINYUSDC.Id)], info[strconv.Itoa(ALGO.Id)])
	fmt.Printf("share of pool: %.3f\n", shareFloat*100)

}
