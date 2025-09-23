# TX based exchange integration

The following documentation is relevant for the Qubic "2.0" RPC.

## Table of contents

## Endpoint summary

The RPC infrastructure is composed of several services that serve different purposes.  
This section briefly describes these services and their endpoints.

> [Full API Swagger Documentation](swagger/qubic-query-doc.html)

> **Note:** You can select to see the documentation for different services by clicking on the `Select a definition`
> dropdown menu on the top-right side of the page.

### Query service

The purpose of the query service is to serve archived data such as tick (block) information, transactions, identity
transactions, etc.

| Method | Endpoint                    | Description                                                                                             |
|--------|-----------------------------|---------------------------------------------------------------------------------------------------------|
| POST   | /getTickData                | Query the data related to a certain tick.                                                               |
| POST   | /getTransactionByHash       | Query the data related to a certain transaction.                                                        |
| POST   | /getTransactionsForIdentity | Query the transactions of a certain identity (address). Allows for different filters and range options. |
| POST   | /getTransactionsForTick     | Query the transactions of a certain tick.                                                               |
| POST   | /getComputorListsForEpoch   | Query the computor list of a certain epoch.                                                             |
| GET    | /getLastProcessedTick       | Retrieve the number of the last archived tick.                                                          |
| GET    | /getProcessedTickIntervals  | Retrieve the archived tick intervals in relation to their epoch.                                        |

### Live service

The live service acts as a proxy to the live network and allows for querying certain information directly from the
network, sending transactions and querying smart contract data.

| Method | Endpoint                  | Description                                                                             |
|--------|---------------------------|-----------------------------------------------------------------------------------------|
| POST   | /v1/broadcast-transaction | Broadcast a new transaction to the network.                                             |
| POST   | /v1/querySmartContract    | Perform a query on a smart contract function.                                           |
| GET    | /v1/tick-info             | Query the current tick of the network.                                                  |
| GET    | /v1/balances/{identity}   | Query the balance of a certain identity, alongside with some transfer related metadata. |

## Qubic workflow guidelines

Integrating with Qubic requires following certain rules and guidelines

### Only one concurrent transaction per source address

Due to Qubic's architecture, only one concurrent transaction from a source address can exist in the network.  
Sample workflow:

1. Network is on tick 5.
2. Address A sends one transaction, scheduled for execution on tick 10.
3. Network is on a tick lower than 10 and address A sends a new transaction scheduled for execution on tick 15. `=> The
   transaction for tick 10 will get overwritten by the transaction for tick 15.`
4. Tick is still on a tick lower than 9 and address A sends another transaction, this time targeting tick 9.
   `=> The transaction for tick 9 will be ignored this time. The network will keep the transaction targeting the higher tick.`

> **Takeaway:** Only one concurrent transaction per source address. Do not send transactions from the same address until
> the tick of the previous transaction has been reached.

### Respect status information

There are a couple of endpoints that can be used to get different types of status information:

- `/v1/tick-info` -> Query the current tick of the network. New transactions must have a target tick larger than this
  value. Note that the network keeps on ticking, thus we recommend setting the target tick of new transaction to 15 - 20
  ticks more than the current network tick.
- `/getLastProcessedTick` -> Query the number of the last archived tick. The archival process is slightly behind the
  network, so **this value does not represent the current state of the network**. You can only query information for
  ticks lower and equal to this value.
- `/getProcessedTickIntervals` -> Query the tick intervals of all the stored epochs. Note that due to different network
  conditions, there may exist multiple intervals for certain epochs, as the network can restart and skip some ticks in
  certain scenarios.

### Epoch transition

Every Wednesday at 12 PM UTC, the Qubic network undergoes a process known as "epoch transition", during which, **the
network is unreachable**.  
This process usually takes from a couple of minutes to an hour.

> **Note that the RPC API is reachable**, but no new transactions can be sent during this time, and no new information
> is
> available until the transition is finished.

During an epoch transition:

1. Network nodes are updated to a new software version.
2. **Data from the previous epoch is discarded**, thus the need for the archival service provided by the integration
   layer.

## General code examples

> The following examples are written in GO and use the [go-node-connector](https://github.com/qubic/go-node-connector)
> library.

### Generating a seed (wallet)

```Go
func generateWallet() error {

seed := types.GenerateRandomSeed()
fmt.Printf("Seed: %s\n", seed)

wallet, err := types.NewWallet(seed)
if err != nil {
return fmt.Errorf("failed to create wallet: %w", err)
}
fmt.Printf("Wallet identity (address): %s\n", wallet.Identity.String())
fmt.Printf("Private key: %s\n", hex.EncodeToString(wallet.PrivKey[:]))
fmt.Printf("Public key: %s\n", hex.EncodeToString(wallet.PubKey[:]))

return nil
}
```

### Creating, signing, sending and verifying a transaction

#### Overview

The basic steps for this process are:

1. Request current network tick from `/v1/tick-info`.
2. Create transaction, define its target tick as a tick in the future and sign it.
3. Send transaction and store its hash.
4. Verify transaction by querying `/getTransactionByHash` after the tick returned by `/getLastProcessedTick` surpasses
   the target tick of the transaction.

> Note: Please do not create another transaction for the same sender address before the previous one has been completed.
**Your transaction may be overwritten.**

#### Example

For this example, it is assumed that you already have the information required to create a simple transaction:

- Sender identity and seed
- Destination identity
- Funds inside the sender wallet.

```Go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/qubic/go-node-connector/types"
)

const baseUrl = `https://api.qubic.org`

func main() {
	err := run()
	if err != nil {
		fmt.Printf("error: %v", err)
	}
}

func run() error {

	// Define sender address, sender seed, destination address and amount
	sourceID := ""
	sourceSeed := ""
	destinationID := ""
	amount := int64(10)

	// Create live service client utility object
	lsc := types.NewLiveServiceClient(baseUrl)

	// Get current network tick
	tickInfoResponse, err := lsc.GetTickInfo()
	if err != nil {
		return fmt.Errorf("getting tick info: %w", err)
	}

	// Define a target tick in the future
	targetTick := tickInfoResponse.TickInfo.Tick + 15

	// Create a simple transaction
	tx, err := types.NewSimpleTransferTransaction(sourceID, destinationID, amount, targetTick)
	if err != nil {
		return fmt.Errorf("creating simple transfer transaction: %w", err)
	}

	// Create signer object based on the sender's seed, then sign the transaction
	signer, err := types.NewSigner(sourceSeed)
	if err != nil {
		return fmt.Errorf("creating signer: %w", err)
	}

	tx, err = signer.SignTx(tx)
	if err != nil {
		return fmt.Errorf("signing transaction: %w", err)
	}

	// Broadcast the transaction and store its hash (id) in a variable
	txBroadcastResponse, err := lsc.BroadcastTransaction(tx)
	if err != nil {
		return fmt.Errorf("broadcasting transaction: %w", err)
	}
	txId := txBroadcastResponse.TransactionId

	fmt.Printf("Broadcast transaction %s. Scheduled for execution on tick %d.\n", txId, targetTick)

	// Query the last processed tick every second until the target tick is reached
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			lastProcessedTick, err := fetchLastProcessedTick()
			if err != nil {
				fmt.Printf("Error fetching last processed tick: %v\n", err)
				continue
			}
			fmt.Printf("Last processed tick: %d\n", lastProcessedTick)
			// If the last processed tick has reached the scheduled tick, we can try to query the transaction and print its data
			if lastProcessedTick >= targetTick {
				err = fetchAndPrintTransactionData(txId)
				if err != nil {
					fmt.Printf("fetching and printing transaction data: %v\n", err)
					continue
				}
				return nil
			}

		}
	}
}

func fetchLastProcessedTick() (uint32, error) {
	request, err := http.NewRequest(http.MethodGet, baseUrl+"/getLastProcessedTick", nil)
	if err != nil {
		return 0, fmt.Errorf("creating last processed tick request: %w", err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return 0, fmt.Errorf("performing last processed tick request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("last processed tick request returned status %d", response.StatusCode)
	}

	var responseObject struct {
		TickNumber uint32 `json:"tickNumber"`
	}

	err = json.NewDecoder(response.Body).Decode(&responseObject)
	if err != nil {
		return 0, fmt.Errorf("decoding last processed tick response: %w", err)
	}

	return responseObject.TickNumber, nil
}

func fetchAndPrintTransactionData(txId string) error {
	payloadObject := struct {
		Hash string `json:"hash"`
	}{
		Hash: txId,
	}
	marshalledPayload, err := json.Marshal(payloadObject)
	if err != nil {
		return fmt.Errorf("marshalling transaction by hash payload: %w", err)
	}

	request, err := http.NewRequest(http.MethodPost, baseUrl+"/getTransactionByHash", bytes.NewReader(marshalledPayload))
	if err != nil {
		return fmt.Errorf("creating transaction by hash request: %w", err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("performing transaction by hash request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {

		// read response body for more details
		var respBody bytes.Buffer
		_, _ = respBody.ReadFrom(response.Body)
		fmt.Printf("Response body: %s\n", respBody.String())

		return fmt.Errorf("transaction by hash request returned status %d", response.StatusCode)
	}

	var responseObject struct {
		Hash        string `json:"hash"`
		Amount      string `json:"amount"`
		Source      string `json:"source"`
		Destination string `json:"destination"`
		TickNumber  uint32 `json:"tickNumber"`
		Timestamp   string `json:"timestamp"`
		InputType   uint32 `json:"inputType"`
		InputSize   uint32 `json:"inputSize"`
		InputData   string `json:"inputData"`
		Signature   string `json:"signature"`
		MoneyFlew   bool   `json:"moneyFlew"`
	}

	err = json.NewDecoder(response.Body).Decode(&responseObject)
	if err != nil {
		return fmt.Errorf("decoding transaction by hash response: %w", err)
	}

	// Print the transaction data
	fmt.Printf("Transaction data: %+v\n", responseObject)
	fmt.Printf("Funds transferred: %t\n", responseObject.MoneyFlew)

	return nil
}

```

## Deposit workflow

We assume that, in you business logic you have a list of the accounts of your clients.  
In order to detect a deposit to a client account, you can use the RPC API to run a sequential tick scan.  
The general process would look something like this:

1. Query the tick transactions, with an amount larger than 0 and moneyFlew status `true`.
2. Iterate over the tick transactions and check if transaction destination identity is one of you clients.
3. Credit client account accordingly.

> In general, we suggest to not allow your clients to use their deposit accounts for smart contract usage (e.g. pool
> payouts, quottery or any future use case).  
> However, there is a send many smart contract case you should support.

The below example handles both normal and send-many deposits.

Example:

```Go
package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/qubic/go-node-connector/types"
)

const baseUrl = `https://api.qubic.org`

func main() {

	err := run()
	if err != nil {
		log.Fatalf("Error: %s", err.Error())
	}

}

func run() error {

	tickNumber := uint32(00000000)

	tickTransactions, err := queryTickTransactions(tickNumber)
	if err != nil {
		return fmt.Errorf("querying tick transactions: %w", err)
	}

	err = creditClientTransactions(tickTransactions)
	if err != nil {
		return fmt.Errorf("crediting client transactions: %w", err)
	}

	return nil
}

type Transaction struct {
	Hash        string `json:"hash"`
	Amount      string `json:"amount"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	TickNumber  uint32 `json:"tickNumber"`
	Timestamp   string `json:"timestamp"`
	InputType   uint32 `json:"inputType"`
	InputSize   uint32 `json:"inputSize"`
	InputData   string `json:"inputData"`
	Signature   string `json:"signature"`
	MoneyFlew   bool   `json:"moneyFlew"`
}

func queryTickTransactions(tickNumber uint32) ([]Transaction, error) {

	payloadObject := struct {
		TickNumber uint32 `json:"tickNumber"`
	}{
		TickNumber: tickNumber,
	}
	marshalledPayload, err := json.Marshal(payloadObject)
	if err != nil {
		return nil, fmt.Errorf("marshalling tick transactions payload: %w", err)
	}
	request, err := http.NewRequest(http.MethodPost, baseUrl+"/getTransactionsForTick", bytes.NewReader(marshalledPayload))
	if err != nil {
		return nil, fmt.Errorf("creating tick transactions request: %w", err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("performing tick transactions request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		// read response body for more details
		var respBody bytes.Buffer
		_, _ = respBody.ReadFrom(response.Body)
		fmt.Printf("Response body: %s\n", respBody.String())
		return nil, fmt.Errorf("tick transactions request returned status %d", response.StatusCode)
	}

	var tickTransactions []Transaction
	err = json.NewDecoder(response.Body).Decode(&tickTransactions)
	if err != nil {
		return nil, fmt.Errorf("decoding tick transactions response: %w", err)
	}

	return tickTransactions, nil
}

func creditClientTransactions(transactions []Transaction) error {

	// Iterate through transaction list and filter for successful non-0 amount transactions for which the destination is a client address or send many SC
	for _, transaction := range transactions {
		amount, err := strconv.ParseInt(transaction.Amount, 10, 0)
		if err != nil {
			return fmt.Errorf("parsing transaction fund amount: %w", err)
		}
		if amount == 0 || transaction.MoneyFlew == false {
			continue
		}

		if isClientAddress(transaction.Destination) {
			fmt.Printf("Direct transfer to client address %s of amount %d\n", transaction.Destination, amount)
			// Business logic to credit client account and transfer to hot wallet
			continue
		}

		// Check if transaction is of send-many type
		if transaction.Destination == types.QutilAddress &&
			transaction.InputType == types.QutilSendManyInputType &&
			transaction.InputSize == types.QutilSendManyInputSize &&
			transaction.InputData != "" {

			decodedInput, err := base64.StdEncoding.DecodeString(transaction.InputData)
			if err != nil {
				return fmt.Errorf("decoding send-many input payload: %w", err)
			}

			var sendManyPayload types.SendManyTransferPayload
			err = sendManyPayload.UnmarshallBinary(decodedInput)
			if err != nil {
				log.Fatalf("got err: %s when unmarshalling payload", err.Error())
			}

			transfers, err := sendManyPayload.GetTransfers()
			if err != nil {
				log.Fatalf("got err: %s when getting transfers", err.Error())
			}

			for _, transfer := range transfers {
				if !isClientAddress(transfer.AddressID.String()) {
					continue
				}

				fmt.Printf("Send-many transfer to client address %s of amount %d\n", transfer.AddressID.String(), transfer.Amount)
				// Business logic to credit client account and transfer to hot wallet
			}

		}

	}

	return nil

}

func isClientAddress(identity string) bool {
	var clientAddresses []string // List of your client addresses
	for _, clientAddress := range clientAddresses {
		if identity == clientAddress {
			return true
		}
	}
	return false
}

```

## Withdraw workflow

Withdrawals can can be performed via plain transactions or the send-many smart contract.

### Plain transactions

Plain transactions are fee-less and limited to one concurrent transaction per hot wallet. The process is described in
the [Creating, signing, sending and verifying a transaction](#creating-signing-sending-and-verifying-a-transaction)
section.

### Qutil (Send Many) smart contract

The Qutil (also known as Send Many) smart contract allows for multiple (up to 25) fund transfers in a single
transaction.  
The requirements for using this smart contract are as follows:

- The transaction must be sent to the Qutil SC address.
- The transfers must be defined in the transaction payload accordingly.
- A small fee must be paid to the SC. This fee is added to the transfer fund sum.
- Sufficient funds must exist in the sender's wallet.

Example

> Note: In this example, the library adds the SC fee to the total amount automatically.

```Go
package main

import (
	"log"

	"github.com/pkg/errors"
	"github.com/qubic/go-node-connector/types"
)

func SendManyTransactionExample() error {

	senderAddress := ""
	senderSeed := ""

	// Create the list of recipients
	transfers := []types.SendManyTransfer{
		{
			AddressID: "AAA...",
			Amount:    10,
		},
		{
			AddressID: "BBB...",
			Amount:    20,
		},
	}

	var payload types.SendManyTransferPayload

	err := payload.AddTransfers(transfers)
	if err != nil {
		return errors.Wrap(err, "adding transfers to send many payload")
	}

	// Create live service client and get current tick / block number
	lsc := types.NewLiveServiceClient("https://api.qubic.org")
	currentTickInfo, err := lsc.GetTickInfo()
	if err != nil {
		return errors.Wrap(err, "getting current tick info")
	}

	// Schedule transaction for a future tick
	targetTick := currentTickInfo.TickInfo.Tick + 15

	// Create transaction
	tx, err := types.NewSendManyTransferTransaction(senderAddress, targetTick, payload)
	if err != nil {
		return errors.Wrap(err, "creating send many transaction")
	}

	// Create signer based on the sender's seed and sign the transaction
	signer, err := types.NewSigner(senderSeed)
	if err != nil {
		return errors.Wrap(err, "creating signer")
	}

	tx, err = signer.SignTx(tx)
	if err != nil {
		return errors.Wrap(err, "signing transaction")
	}

	// Broadcast the transaction
	response, err := lsc.BroadcastTransaction(tx)
	if err != nil {
		return errors.Wrap(err, "broadcasting transaction")
	}

	log.Printf("Broadcasted transaction '%s' to %d peers. Scheduled for tick %d\n", response.TransactionId, response.PeersBroadcasted, targetTick)

	return nil
}

```

## Asset transfers

In order to transfer assets, you can refer to this example:

```Go
package main

import (
	"log"

	"github.com/pkg/errors"
	"github.com/qubic/go-node-connector/types"
)

func AssetTransferTransactionExample() error {

	senderAddress := ""
	senderSeed := ""

	destinationAddress := ""

	assetName := ""
	assetIssuer := ""
	numberOfUnits := int64(1)

	// The transfer fee may be subject to change in the future.
	transferFee := int64(100)

	// Create the asset transfer payload
	payload, err := types.NewAssetTransferPayload(assetName, assetIssuer, destinationAddress, numberOfUnits)
	if err != nil {
		return errors.Wrap(err, "creating asset transfer payload")
	}

	// Create live service client and get current tick / block number
	lsc := types.NewLiveServiceClient("https://api.qubic.org")
	currentTickInfo, err := lsc.GetTickInfo()
	if err != nil {
		return errors.Wrap(err, "getting current tick info")
	}

	// Schedule transaction for a future tick
	targetTick := currentTickInfo.TickInfo.Tick + 15

	// Create transaction
	tx, err := types.NewAssetTransferTransaction(senderAddress, targetTick, transferFee, payload)
	if err != nil {
		return errors.Wrap(err, "creating asset transfer transaction")
	}

	// Create signer based on the sender's seed and sign the transaction
	signer, err := types.NewSigner(senderSeed)
	if err != nil {
		return errors.Wrap(err, "creating signer")
	}

	tx, err = signer.SignTx(tx)
	if err != nil {
		return errors.Wrap(err, "signing transaction")
	}

	// Broadcast the transaction
	response, err := lsc.BroadcastTransaction(tx)
	if err != nil {
		return errors.Wrap(err, "broadcasting transaction")
	}

	log.Printf("Broadcasted transaction '%s' to %d peers. Scheduled for tick %d\n", response.TransactionId, response.PeersBroadcasted, targetTick)

	return nil
}
```

