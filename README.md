# Tinyman Mobile SDK

Tinyman go SDK compatible with go-mobile package.



**Currently Under Testing**



# Go-Mobile Compatibility

Due to go-mobile type restrictions for exported symbols:

- Numeric values that aren't likely to have a value more than 64 bit have an`int` type.
- Numeric Values like Balance that may have a large value have a `string` type and during mathmatical calculations, they will be converted to `*big.Int` or `*big.Float` type.
- For methods that return a map( like `FetchExcessAmounts` method ), there is a method that ends with `Str` ( like `FetchExcessAmountsStr` ) that return JSON String of the map.
- Struct fields that are not supported by go-mobile are unexported and there is a getter method for each one of them in the form of `GetFieldName`



# Preview



```go
import (
	"github.com/soheil555/tinyman-mobile-sdk/v1/client"
   	"github.com/algorand/go-algorand-sdk/mnemonic"
    "github.com/algorand/go-algorand-sdk/crypto"
    "github.com/kr/pretty"
)


privateKey, err := mnemonic.ToPrivateKey("MNEMONIC")

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

// Get a quote for a swap of 1 ALGO to TINYUSDC with 1% slippage tolerance
quote, err := pool.FetchFixedInputSwapQuote(ALGO.Call("1000000"), 0.01)
if err != nil {
    fmt.Printf("error fetching fixed input swap quote: %s\n", err)
    return
}

pretty.Println(quote)
```





# Examples



### Basic Swapping

[swapping1](https://github.com/soheil555/tinyman-mobile-sdk/blob/main/examples/swapping1/main.go) This example demonstrates basic functionality including:

- retrieving Pool details
- getting a swap quote
- preparing swap transactions
- signing transactions
- submitting transactions
- checking excess amounts
- preparing redeem transactions



### Basic Pooling

[pooling1](https://github.com/soheil555/tinyman-mobile-sdk/blob/main/examples/pooling1/main.go) This example demonstrates retrieving the current pool position/share for an address.



### Basic Add Liquidity (Minting)

[add-liquidity](https://github.com/soheil555/tinyman-mobile-sdk/blob/main/examples/add-liquidity1/main.go) This example demonstrates add liquidity to an existing pool.





## Running Examples

1. inside project base directory

   ```bash
   cp .env.example .env
   ```

2. you need to open `.env` file and set the **MNEMONIC** variable

3. change directory to examples/[example-name]

4. run the following command to start running the example

   ```bash
   go run main.go
   ```






# Running Tests

In the project base directory

```bash
go test -v ./...
```





# Conventions



- Methods starting with `Fetch` all make network requests to fetch current balances/state.
- Methods of the form `PrepareXTransactions` all return `TransactionGroup` structs (see below).
- All asset amounts are returned as `AssetAmount` structs which contain an `Asset` and `amount` (`string`).
- All asset amount inputs are expected as micro units e.g. 1 Algo = 1_000_000 micro units.




# License

Available under the MIT license. See the `LICENSE` file for more info.

