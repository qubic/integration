# TX Based Exchange integration
The following use cases are based on the Qubic RPC V1.

The following examples refer to the Qubic RPC V1 API and the [TS Library](https://github.com/qubic/ts-library)

The TS Library uses a Webassembly for all the cryptographical stuff.

For all the following examples the `baseUrl` is set to:
```js
const baseUrl = 'https://testapi.qubic.org/v1';
```

This documentation refers to the [Qubic V1 RPC API](qubic-rpc-doc.html).

| Method  	| Endpoint    	                                | Description   	                                               |  Body Payload |
|---	|----------------------------------------------|---------------------------------------------------------------|---|
| GET  	| /latestTick   	                              | Get the current tick (block height)   	                       | -   |
| POST  | /broadcast-transaction	                      | Broadcast a transaction    	                                  | `{ "encodedTransaction": "<BASE64RAWTX>" }  `  |
| GET  	| /ticks/{tickNumber}/approved-transactions  	 | Get a List of approved transactions for the given tick 	      |   - |
| GET  	| /tx-status/{txId}  	                         | Get the status of a single transaction 	                      |   - |
| GET  	| /ticks/{tickNumber}/tick-data  	             | Get tick information like timestamp, epoch, included tx ids 	 |   - |
| GET  	| /balances/{addressID}	                       | Get balance for specified address ID	                         |   - |
| GET  	| /status  	                                   | Get the RPC status 	                                          |   - |



## Table of Contents
- [TX Based Exchange integration](#tx-based-exchange-integration)
  - [Table of Contents](#table-of-contents)
  - [Qubic Rules](#qubic-rules)
    - [Rule 1: One concurent TX per source Address](#rule-1-one-concurent-tx-per-source-address)
    - [Rule 2: Respect RPC Status](#rule-2-respect-rpc-status)
    - [Rule 3: Epoch change](#rule-3-epoch-change)
  - [General Examples](#general-examples)
    - [Generating a Seed](#generating-a-seed)
    - [Signing a Package](#signing-a-package)
    - [Create, sign, send and verify a transaction](#create-sign-send-and-verify-a-transaction)
      - [Step 1: Request the latest tick height](#step-1-request-the-latest-tick-height)
      - [Step 2: Create and sign transaction](#step-2-create-and-sign-transaction)
      - [Step 3: Send transaction](#step-3-send-transaction)
      - [Step 4: Verify transaction status](#step-4-verify-transaction-status)
  - [Deposit Workflow](#deposit-workflow)
    - [Scan Ticks/Blocks sequentially](#scan-ticksblocks-sequentially)
      - [Special case Qutil(Send Many) Smart Contract](#special-case-qutilsend-many-smart-contract)
  - [Withdraw Workflow](#withdraw-workflow)
    - [Plain Transaction](#plain-transaction)
    - [Qutil(Send Many) Smart Contract](#qutilsend-many-smart-contract)
  - [Error Handling](#error-handling)


## Qubic Rules
When using the Qubic RPC you should follow some important rules.

### Rule 1: One concurent TX per source Address
Due to Qubics architecture, only one concurrent transaction from a source address can exist in the network.

Sample pseudo workflow:
```
Pre-condition: For the sake of simplicity in this explanation, it is assumed that Alice has a single source address.

1. Network is in tick 5.
2. Alice sends one transaction in the network for tick 10.
3. Network is still in a tick lower than 10 and Alice sends a new transaction in the network for tick 15.
4. The transaction for tick 10 will be overwritten with the transaction for tick 15. In order to not ovewrite the transaction for tick 10, Alice should have waited for the tick 10 to pass before to send the new transaction.
5. Network is still in a tick lower than 9 and Alice sends a new transaction for tick 9.
6. The transaction sent for tick 9 will be ignored. The transaction with higher tick (the one for tick 15 in this example) will be kept in the network.

```

> [!IMPORTANT]
> Only one concurrent TX per sending ID, don’t send a new transaction from the same source address until the previous tx target tick has been passed.



### Rule 2: Respect RPC Status
The endpoint
- `/status`

returns the current status of the RPC data.

Example response can be found below:

```json
{
    "lastProcessedTick": {
        "tickNumber": 13189868,
        "epoch": 102
    },
    "lastProcessedTicksPerEpoch": {
        "99": 12941678,
        "100": 13059142,
        "101": 13101159,
        "102": 13189868
    },
    "skippedTicks": [
        {
            "startTick": 12941679,
            "endTick": 12949999
        },
        {
            "startTick": 13059143,
            "endTick": 13059999
        },
        {
            "startTick": 13101160,
            "endTick": 13109999
        }
    ],
    "processedTickIntervalsPerEpoch": [
        {
            "epoch": 101,
            "intervals": [
                {
                    "initialProcessedTick": 13082191,
                    "lastProcessedTick": 13101159
                }
            ]
        },
        {
            "epoch": 102,
            "intervals": [
                {
                    "initialProcessedTick": 13110000,
                    "lastProcessedTick": 13189868
                }
            ]
        }
    ]
}
```

> [!IMPORTANT]
> The `lastProcessedTick` indicates up to which tick the RPC server has processed data.

> In the array `processedTickIntervalsPerEpoch` you find the processed ticks. Due to updates, epochs might have multiple tick ranges.

### Rule 3: Epoch change
Every Wednesday, Qubic has its epoch change. The epoch transition is at ~12:00 UTC.
If the network gets a breaking update, the epoch transition may take a few minutes. Expect between 11:59 to 13:00 UTC a delay for transactions.

## General Examples

### Generating a Seed
A seed is the private key in Qubic. Based on the seed, you can create the public address (id).

The seed is a 55-lower-case-char string. Please use a proper random generator in your environment. The code below is for demonstration purposes only.

**Javascript**
```js
    // generates a random seed
  seedGen() {
    const letters = "abcdefghijklmnopqrstuvwxyz";
    const letterSize = letters.length;
    let seed = "";
    for (let i = 0; i < 55; i++) {
      seed += letters[Math.floor(Math.random() * letterSize)];
    }
    return seed;
  }

  // get a reference to the helper class
  const helper = new QubicHelper();

  // generate a seed
  const seed = seedGen();

  // generate the id package
  const {privateKey, publicKey, publicId} = await helper.createIdPackage(seed);

  // the resulting package contains:
  // privateKey => your binary private Key; this can be used to sign packages (e.g. transactions)
  // publicKey => the public key
  // publicId => the public key in human readable format. this is the address qubic users use
```

**Go**
```go
package main

import (
	"encoding/hex"
	"fmt"
	"github.com/qubic/go-node-connector/types"
	"log"
)

func main() {
	seed := types.GenerateRandomSeed()
	fmt.Println(seed)
	wallet, err := types.NewWallet(seed)
	if err != nil {
		log.Fatalf("got err: %s when creating wallet", err.Error())
	}

	fmt.Println(wallet.Identity.String())
	fmt.Println(hex.EncodeToString(wallet.PrivKey[:]))
	fmt.Println(hex.EncodeToString(wallet.PubKey[:]))
}
```

### Signing a Package
For signing a package, you can either use the qubic crypto library or e.g. for transactions the wrapper from the `ts-library`.

The pre-condition to be able to sign a package is to have the `seed` or `privateKey`.

The following example assumes that we have already created our `idPackage` which includes our `privateKey`.

**Javascript**
```js
  // to sign a package, you need the private key which is derived from the seed and its publicKey
  const seed = 'wqbdupxgcaimwdsnchitjmsplzclkqokhadgehdxqogeeiovzvadstt';
  const idPackage = await helper.createIdPackage(seed);

  // fake package!
  const packet = new Uint8Array(32);

  // example without await, you can also use async/await
  const signedPacket = await crypto.then(({ schnorrq, K12 }) => {
            
            // create the digest store; digest length is defined by 32 bytes
            const digest = new Uint8Array(QubicDefinitions.DIGEST_LENGTH);
            
            // you may receive packet in a function
            const toSignPacket = packet;

            // create K12 digest
            K12(toSignPacket, digest, QubicDefinitions.DIGEST_LENGTH);

            // sign the packet and receive signature
            // the signing needs the inputs:
            // privateKey
            // publicKey
            // K12 digest of the packet to sign
            const signature = schnorrq.sign(idPackage.privateKey, idPackage.publicKey, digest);

            // normally you would add the signature to the packet for transfer
            var signedData = new Uint8Array(toSignPacket.length + signature.length);
            signedData.set(toSignPacket);
            signedData.set(signature, toSignPacket.length);
            
            // after combining packet + signature you would take another digest
            // this new digest can be considered as the id of the complete package (e.g. transaction id)
            K12(signedData, digest, QubicDefinitions.DIGEST_LENGTH)
            
            return {
                signedData: signedData,
                digest: digest,
                signature: signature
            };
        });
```


**Go**
```go
package main

import (
  "encoding/hex"
  "fmt"
  "github.com/qubic/go-schnorrq"
  "github.com/qubic/go-node-connector/types"
  "log"
)

const baseUrl = "https://testapi.qubic.org/v1"

func main() {
  tx, err := types.NewSimpleTransferTransaction(
    "source-addr-here",
    "dest-addr-here",
    1,
    13255850,
  )
  if err != nil {
    log.Fatalf("got err: %s when creating simple transfer transaction", err.Error())
  }
  fmt.Printf("source pubkey: %s\n", hex.EncodeToString(tx.SourcePublicKey[:]))

  unsignedDigest, err := tx.GetUnsignedDigest()
  if err != nil {
    log.Fatalf("got err: %s when getting unsigned digest local", err.Error())
  }

  subSeed, err :=types.GetSubSeed("seed-here")
  if err != nil {
    log.Fatalf("got err: %s when getting subSeed", err.Error())
  }
  
  sig, err := schnorrq.Sign(subSeed, tx.SourcePublicKey, unsignedDigest)
  if err != nil {
    log.Fatalf("got err: %s when signing", err.Error())
  }
  fmt.Printf("sig: %s\n", hex.EncodeToString(sig[:]))
  tx.Signature = sig

  encodedTx, err := tx.EncodeToBase64()
  if err != nil {
    log.Fatalf("got err: %s when encoding tx to base 64", err.Error())
  }
  fmt.Printf("encodedTx: %s\n", encodedTx)

  id, err := tx.ID()
  if err != nil {
    log.Fatalf("got err: %s when getting tx id", err.Error())
  }

  fmt.Printf("tx id(hash): %s\n", id)
}
```
Please find a complete example of transaction signing here: https://github.com/qubic/ts-library/blob/main/test/createTransactionTest.js. The complete source code can be found in the same repo.

### Create, sign, send and verify a transaction
#### Process overview
The basic steps for this process are:
- request `latestTick` from `/latestTick` endpoint
- create and sign transaction
- send transaction with `latestTick + 5` (= your `targetTick`), store tx hash
- to verify, either:
 - poll `/ticks/{tickNumber}/approved-transactions` until it returns 200 for your `targetTick`. If the tx hash is available there, the tx was successful (as in examples down below), OR:
 - poll `/status` for `lastProcessedTick` and as soon as `lastProcessedTick` > `targetTick`, check `/ticks/{tickNumber}/approved-transactions`. If the tx hash is available there, the tx was successful OR:
 - poll `/status` for `lastProcessedTick` and as soon as `lastProcessedTick` > `targetTick`, check `/tx-status/{txId}`. If `moneyFlew = true`, the transaction was successful
- **only after those steps you may create another transaction as you will overwrite an existing one if sent to the network before the target tick has passed.**

For the following examples of this flow, we assume you have already all needed data to create and send the transaction:
- Sender (including seed)
- Receiver
- Amount


#### Step 1: Request the latest tick height

**Javascript**
```js
const response = await fetch(`${baseUrl}/latestTick`);
const tickResponse = await response.json();
const latestTick = tickResponse.latestTick;

// don't forget to do proper error handling!

```

**Go**
```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

const baseUrl = "https://testapi.qubic.org/v1"

func main() {
	url := baseUrl + "/latestTick"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatalf("got err: %s when creating request", err.Error())
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("got err: %s when performing request", err.Error())
	}
	defer res.Body.Close()

    if res.StatusCode != http.StatusOK {
      log.Fatalf("Got non 200 status code: %d", res.StatusCode)
    }

	type response struct {
		LatestTick uint32 `json:"latestTick"`
	}
	var body response

	err = json.NewDecoder(res.Body).Decode(&body)
	if err != nil {
		log.Fatalf("got err: %s when decoding body", err.Error())
	}

	fmt.Println(body.LatestTick)
}
```

#### Step 2: Create and sign transaction

**Javascript**
```js
  // please find an extended example here: https://github.com/qubic/ts-library/blob/main/test/createTransactionTest.js

  // create and sign transaction
  const tx = new QubicTransaction().setSourcePublicKey(sourcePublicKey)
      .setDestinationPublicKey(destinationPublicKey)
      .setAmount(amount)
      // it is important to set a target tick for the execution of this transaction
      // a suggested offset is 5-7 ticks
      .setTick(latestTick + 5);

    // will build the tx: bundle all values and sign it
    // returns the raw bytes of signed transaction
    const signedTransactionData = await tx.build(signSeed);

    // by requesting getId() you receive the txId of this transaction
    // this id can presented to the client as reference
    const transactionId = tx.getId();
```
**Go**
```go
package main

import (
  "bytes"
  "encoding/hex"
  "encoding/json"
  "fmt"
  "github.com/qubic/go-schnorrq"
  "github.com/qubic/go-node-connector/types"
  "log"
  "net/http"
)

const baseUrl = "https://testapi.qubic.org/v1"

func main() {
  tx, err := types.NewSimpleTransferTransaction(
    "source-addr-here",
    "dest-addr-here",
    1,
    13255850,
  )
  if err != nil {
    log.Fatalf("got err: %s when creating simple transfer transaction", err.Error())
  }
  fmt.Printf("source pubkey: %s\n", hex.EncodeToString(tx.SourcePublicKey[:]))

  unsignedDigest, err := tx.GetUnsignedDigest()
  if err != nil {
    log.Fatalf("got err: %s when getting unsigned digest local", err.Error())
  }

  subSeed, err :=types.GetSubSeed("seed-here")
  if err!= nil {
    log.Fatalf("got err %s when getting subSeed", err.Error())
  }
  
  sig, err := schnorrq.Sign(subSeed, tx.SourcePublicKey, unsignedDigest)
  if err != nil {
    log.Fatalf("got err: %s when signing", err.Error())
  }
  fmt.Printf("sig: %s\n", hex.EncodeToString(sig[:]))
  tx.Signature = sig

  encodedTx, err := tx.EncodeToBase64()
  if err != nil {
    log.Fatalf("got err: %s when encoding tx to base 64", err.Error())
  }
  fmt.Printf("encodedTx: %s\n", encodedTx)

  id, err := tx.ID()
  if err != nil {
    log.Fatalf("got err: %s when getting tx id", err.Error())
  }

  fmt.Printf("tx id(hash): %s\n", id)

  url := baseUrl + "/broadcast-transaction"
  payload := struct {
    EncodedTransaction string `json:"encodedTransaction"`
  }{
    EncodedTransaction: encodedTx,
  }

  buff := new(bytes.Buffer)
  err = json.NewEncoder(buff).Encode(payload)
  if err != nil {
    log.Fatalf("got err: %s when encoding payload", err.Error())
  }

  req, err := http.NewRequest(http.MethodPost, url, buff)
  if err != nil {
    log.Fatalf("got err: %s when creating request", err.Error())
  }

  res, err := http.DefaultClient.Do(req)
  if err != nil {
    log.Fatalf("got err: %s when performing request", err.Error())
  }
  defer res.Body.Close()

  if res.StatusCode != http.StatusOK {
    log.Fatalf("Got non 200 status code: %d", res.StatusCode)
  }

  type response struct {
    PeersBroadcasted   uint32 `json:"peersBroadcasted"`
    EncodedTransaction string `json:"encodedTransaction"`
  }
  var body response

  err = json.NewDecoder(res.Body).Decode(&body)
  if err != nil {
    log.Fatalf("got err: %s when decoding body", err.Error())
  }

  fmt.Printf("%+v\n", body)
}

```
#### Step 3: Send transaction

**Javascript**
```js
  // after creating and signing the tx it should be sent to the network
  const response = fetch(`${baseUrl}/broadcast-transaction`,
                    {
                        headers: {
                          'Accept': 'application/json',
                          'Content-Type': 'application/json'
                        },
                        method: "POST",
                        body: JSON.stringify({
                          encodedTransaction: signedTransactionData
                        })
                    });

  if(reponse.status == 200)
  {
    // yipiii! transaction has been broadcasted
  }else {
    // :( something went wrong, try again
  }
```

**Go**
```go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
    "github.com/qubic/go-node-connector/types"
)

const baseUrl = "https://testapi.qubic.org/v1"

func main() {
	var tx types.Transaction
	encodedTransaction, _ = tx.EncodeToBase64()
	url := baseUrl + "/broadcast-transaction"
	payload := struct {
		EncodedTransaction string `json:"encodedTransaction"`
	}{
		EncodedTransaction: encodedTransaction,
	}

	buff := new(bytes.Buffer)
	err := json.NewEncoder(buff).Encode(payload)
	if err != nil {
		log.Fatalf("got err: %s when encoding payload", err.Error())
	}

	req, err := http.NewRequest(http.MethodPost, url, buff)
	if err != nil {
		log.Fatalf("got err: %s when creating request", err.Error())
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("got err: %s when performing request", err.Error())
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Fatalf("Got non 200 status code: %d", res.StatusCode)
	}

	type response struct {
		PeersBroadcasted uint32 `json:"peersBroadcasted"`
	}
	var body response

	err = json.NewDecoder(res.Body).Decode(&body)
	if err != nil {
		log.Fatalf("got err: %s when decoding body", err.Error())
	}

	fmt.Println(body.PeersBroadcasted)
}
```

#### Step 4: Verify transaction status

**Javascript**
```js
  // you can verify if a transaction was successful as soon the target tick has passed (true finality)

   const response = await fetch(`${baseUrl}/ticks/${tx.tick}/approved-transactions`);
      // only if status == 200 (ok) the tx can be marked as verified
    if(response.status == 200) 
    {
      const tickTransactions = await response.json();
      const txStatus = tickTransactions.transactionsStatus.find(f => f.txId === tx.getId());
      if(txStatus){
        // yipiii! transaction was successful
      }else {
        // :( transaction was NOT successful
      }
    }else if(response.status === 400){
      // bad request must be handled by the client. check code table below
    }else if(response.status === 404){
      const errorStatus = await response.json();
      if(errorStatus.code === 123){
        // your request was to early, please repeat it
        // the lastProcessedTick is lower than the one you have requested
      }
    }

```

**Go**
```go
package main

import (
  "encoding/json"
  "fmt"
  "log"
  "net/http"
)

const baseUrl = "https://testapi.qubic.org/v1"

func main() {
  targetTick := 13201784
  url := fmt.Sprintf("%s/ticks/%d/approved-transactions", baseUrl, targetTick)

  req, err := http.NewRequest(http.MethodGet, url, nil)
  if err != nil {
    log.Fatalf("got err: %s when creating request", err.Error())
  }

  res, err := http.DefaultClient.Do(req)
  if err != nil {
    log.Fatalf("got err: %s when performing request", err.Error())
  }
  defer res.Body.Close()

  if res.StatusCode == http.StatusOK {
    type response struct {
      ApprovedTransactions []struct {
        SourceId     string `json:"sourceId"`
        DestId       string `json:"destId"`
        Amount       string `json:"amount"`
        TickNumber   uint32 `json:"tickNumber"`
        InputType    int    `json:"inputType"`
        InputSize    int    `json:"inputSize"`
        InputHex     string `json:"inputHex"`
        SignatureHex string `json:"signatureHex"`
        TxId         string `json:"txId"`
      } `json:"approvedTransactions"`
    }
    var body response

    err = json.NewDecoder(res.Body).Decode(&body)
    if err != nil {
      log.Fatalf("got err: %s when decoding body", err.Error())
    }

    fmt.Printf("%+v", body.ApprovedTransactions)
  } else if res.StatusCode == http.StatusBadRequest {
    type errResponse struct {
      Code    int                      `json:"code"`
      Message string                   `json:"message"`
      Details []map[string]interface{} `json:"details"`
    }
    var body errResponse

    err = json.NewDecoder(res.Body).Decode(&body)
    if err != nil {
      log.Fatalf("got err: %s when decoding error body", err.Error())
    }

    switch body.Code {
    case 9:
      lastProcessedTick := body.Details[0]["lastProcessedTick"]
      fmt.Println(uint32((lastProcessedTick).(float64)))
    case 11:
      nextTickNumber := body.Details[0]["nextTickNumber"]
      fmt.Println(uint32((nextTickNumber).(float64)))
    }
  } else {
    //handle error
  }
}
```

When requesting the RPC server for `approved-transactions` you might receive a `400 Bad Request`.

`Bad Request` example response:
```json
{
    "code": 11,
    "message": "provided tick number 13055400 was skipped by the system, next available tick is 13082191",
    "details": [
        {
            "@type": "type.googleapis.com/qubic.txstatus.pb.NextAvailableTick",
            "nextTickNumber": 13082191
        }
    ]
}
```


## Deposit Workflow

We assume that you have in your business logic the accounts of your clients. We refer to these accounts as `clientAcccount`. 

A client Account is a package containing `seed`, `privateKey`, `publicKey` and `publicId`. 

The list of all `clientAccount` is called `clientAccountList`.


### Scan Ticks/Blocks sequentially

To detect a deposit to a `clientAccount` we use the Qubic RPC and run a sequential tick/blockscan.
You will need to define an initial tick from which on you will start your tick scans. In our example example below, we start with the tick `13032965`.

The following code samples contain pseudo code which you have to replace by your own business logic.

**Javascript**
```js
  // don't forget to do a proper error handling!
  // if you request a tick which is yet not processed, you will receive a 404 with a specific message

  const currentTick = 13032965;

  // request transactions for the tick
  const response = await fetch(`${baseUrl}/ticks/${currentTick}/approved-transactions`);
  
  // the result will contain all executed and approved transactions for the given tick
  const tickTransactions = await response.json();

  // map the transactions to your `clientAccountList`
  const clientDeposits = clientAccountList.filter(f => tickTransactions.find(t => t.destId == f.publicId))

  clientDeposits.forEach(clientAccount => {
    // add transaction to your accounting
    const internalTx = clientAccount.addIncomingTransactions(tickTransactions.filter(f => f.destId == clientAccount.publicId).map(m => createInternalTransaction(m)));

    // start here your internal business logic to credit your clients account
    creditClientAccount(clientAccount, internalTx);

    // start transferring the deposit to your hot wallet here
    transferToHotWallet(clientAccount, internalTx);
  });
```

**Go**
```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

const baseUrl = "https://testapi.qubic.org/v1"

func main() {
	targetTick := 13201784
	url := fmt.Sprintf("%s/ticks/%d/approved-transactions", baseUrl, targetTick)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatalf("got err: %s when creating request", err.Error())
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("got err: %s when performing request", err.Error())
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Fatalf("got non 200 status code: %d", res.StatusCode)
	}

	type response struct {
		ApprovedTransactions []struct {
			SourceId     string `json:"sourceId"`
			DestId       string `json:"destId"`
			Amount       string `json:"amount"`
			TickNumber   uint32 `json:"tickNumber"`
			InputType    int    `json:"inputType"`
			InputSize    int    `json:"inputSize"`
			InputHex     string `json:"inputHex"`
			SignatureHex string `json:"signatureHex"`
			TxId         string `json:"txId"`
		} `json:"approvedTransactions"`
	}
	var body response

	err = json.NewDecoder(res.Body).Decode(&body)
	if err != nil {
		log.Fatalf("got err: %s when decoding body", err.Error())
	}
	
	for _, approvedTx := range body.ApprovedTransactions {
		if !isClientAddress(approvedTx.DestId) {
			continue
		}
		
		//insert business logic to credit client account and transfer to hot wallet
	}

	fmt.Printf("%+v", body.ApprovedTransactions)
}

func isClientAddress(addr string) bool {
	clientAddresses := []string {
		"a", "b", "c",
	}
	
	for _, clientAddr := range clientAddresses {
		if clientAddr == addr {
			return true
		}
	}
	
	return false
}

```

Repeat the code above as long you don't get a `400 Bad Request`.

#### Special case Qutil(Send Many) Smart Contract
In general, we suggest to not allow your clients to use their deposit accounts for smart contract usage (e.g. pool payouts, quottery or any future use case).

However, there is a send many smart contract case you should support. Such a transaction can be identified as follows:

**Javascript**
```js
  // we assume you have in hand a tx object (e.g. from the approved-transactions endpoint)
  const tx = getTxFromApprovedTransactionsEndpoint();

  
  if(
    // address of Qutil(Send Many) SC
    tx.destId === 'EAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAVWRF'
    &&
    // type must be 1
    tx.inputType == 1
    &&
    // size must be 1000
    tx.inputSize = 1000
    &&
    // input must be present
    tx.inputHex
    )
    {
      // if we are here, this tx was a send many sc invovation
      // in the input we have potentially 25 inlay transactions

      // translate input to transfers
      const parsedSendManyPayload = await new QubicTransferSendManyPayload().parse(newPayload).getTransfers();

      // get all client accounts which have got a tx
      const clientDeposits = clientAccountList.filter(f => sendManyTransfers.find(t => t.destId == f.publicId))

      clientDeposits.forEach(clientAccount => {

        // add transaction to your accounting
        const internalTx = clientAccount.addIncomingTransactions(sendManyTransfers.filter(f => f.destId == clientAccount.publicId).map(m => createInternalTransaction(m)));

        // start here your internal business logic to credit your clients account
        creditClientAccount(clientAccount, internalTx);

        // start transferring the deposit to your hot wallet here
        transferToHotWallet(clientAccount, internalTx);
      });>

      
    }

    

```

**Go**
```go
package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/qubic/go-node-connector/types"
	"log"
	"net/http"
)

const baseUrl = "https://testapi.qubic.org/v1"

func main() {
	url := baseUrl + "/transactions/ereivwmylkkjicicsljuyhrhdgneoibmyyoxjwffobriggerszqoiudgtksm"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatalf("got err: %s when creating request", err.Error())
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("got err: %s when performing request", err.Error())
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Fatalf("Got non 200 status code: %d", res.StatusCode)
	}

	type response struct {
		Transaction struct {
			SourceId     string `json:"sourceId"`
			DestId       string `json:"destId"`
			Amount       string `json:"amount"`
			TickNumber   int    `json:"tickNumber"`
			InputType    int    `json:"inputType"`
			InputSize    int    `json:"inputSize"`
			InputHex     string `json:"inputHex"`
			SignatureHex string `json:"signatureHex"`
			TxId         string `json:"txId"`
		} `json:"transaction"`
	}
	var body response

	err = json.NewDecoder(res.Body).Decode(&body)
	if err != nil {
		log.Fatalf("got err: %s when decoding body", err.Error())
	}

	// check if tx dest addr is qutil and input type is 1 (send many) and size and input matches
	if body.Transaction.DestId == types.QutilAddress &&
		body.Transaction.InputType == 1 &&
		body.Transaction.InputSize == types.QutilSendManyInputSize &&
		body.Transaction.InputHex != "" {

		decodedInput, err := hex.DecodeString(body.Transaction.InputHex)
		if err != nil {
			log.Fatalf("got err: %s when decoding input", err.Error())
		}
		var sendmanypayload types.SendManyTransferPayload
		err = sendmanypayload.UnmarshallBinary(decodedInput)
		if err != nil {
			log.Fatalf("got err: %s when unmarshalling payload", err.Error())
		}

		transfers, err := sendmanypayload.GetTransfers()
		if err != nil {
			log.Fatalf("got err: %s when getting transfers", err.Error())
		}

		for _, transfer := range transfers {
			if !isClientAddress(transfer.AddressID.String()) {
				continue
			}

			//insert business logic to credit client account and transfer to hot wallet
		}

		fmt.Printf("transfers: %+v\n", transfers)
	}

}

func isClientAddress(addr string) bool {
	clientAddresses := []string{
		"a", "b", "c",
	}

	for _, clientAddr := range clientAddresses {
		if clientAddr == addr {
			return true
		}
	}

	return false
}

```


## Withdraw Workflow
To do withdraws you can either use a plain transaction or the send many smart contract.

### Plain Transaction
This transaction is feeless and limited to one transaction per hot wallet. 
Follow the process from [Create, sign, send and verify a transaction](#create-sign-send-and-verify-a-transaction)

### Qutil(Send Many) Smart Contract
The fee for using the Smart Contract is `10` Qubic and allows to withdraw to up to 25 clients in a single transaction.

A send many smart contract invocation is a qubic transaction with some specific settings.

The example below shows how to use it.

**Javascript**
```js

  // create the builder package
  const sendManyPayload = new QubicTransferSendManyPayload();

  // add a destination
  sendManyPayload.addTransfer({
    destId: new PublicKey("SUZFFQSCVPHYYBDCQODEMFAOKRJDDDIRJFFIWFLRDDJQRPKMJNOCSSKHXHGK"),
    amount: new Long(1)
  });

  // add a destination
  sendManyPayload.addTransfer({
    destId: new PublicKey("SUZFFQSCVPHYYBDCQODEMFAOKRJDDDIRJFFIWFLRDDJQRPKMJNOCSSKHXHGK"),
    amount: new Long(2)
  });

  // ...add up to 25 destination addresses

  // add the fixed fee to the total amount
  const totalAmount = sendManyPayload.getTotalAmount() + BigInt(QubicDefinitions.QUTIL_SENDMANY_FEE);

  // build and sign tx
  const tx = new QubicTransaction().setSourcePublicKey(sourcePublicKey)
    .setDestinationPublicKey(QubicDefinitions.QUTIL_ADDRESS) // a send many transfer should go the Qutil SC
    .setAmount(totalAmount) // calculated from all transfers + fee
    .setTick(0) // set an appropriate target tick
    .setInputType(QubicDefinitions.QUTIL_SENDMANY_INPUT_TYPE) // input type for send many invocation
    .setInputSize(sendManyPayload.getPackageSize()) // the input size equals the size of the send many payload
    .setPayload(sendManyPayload); // add payload

  const signedTransactionData = await tx.build(signSeed);
  
  // the tx can now be sent
 const response = fetch(`${baseUrl}/broadcast-transaction`,
                    {
                        headers: {
                          'Accept': 'application/json',
                          'Content-Type': 'application/json'
                        },
                        method: "POST",
                        body: JSON.stringify({
                          encodedTransaction: signedTransactionData
                        })
                    });

  if(reponse.status == 200)
  {
    // yipiii! transaction has been broadcasted

    // by requesting getId() you receive the txId of this transaction
    // this id can presented to the client as reference
    const transactionId = tx.getId();

    // after a tx has been send you can either check its status by tick scan or by dedicated call to /tx-status/<TDID>

  }else {
    // :( something went wrong, try again
  }


```
**Go**
```go
package main

import (
  "bytes"
  "encoding/hex"
  "encoding/json"
  "fmt"
  "github.com/qubic/go-schnorrq"
  "github.com/qubic/go-node-connector/types"
  "log"
  "net/http"
)

const baseUrl = "https://testapi.qubic.org/v1"

func main() {
  // max 25
  transfers := []types.SendManyTransfer{
    {
      AddressID: "dest-addr-1",
      Amount:    1,
    },
    {
      AddressID: "dest-addr-2",
      Amount:    2,
    },
  }
  var sendManyPayload types.SendManyTransferPayload

  err := sendManyPayload.AddTransfers(transfers)
  if err != nil {
    log.Fatalf("got err: %s when adding transfers", err.Error())
  }
  tx, err := types.NewSendManyTransferTransaction(
    "source-addr",
    13256050,
    sendManyPayload,
  )
  if err != nil {
    log.Fatalf("got err: %s when creating simple transfer transaction", err.Error())
  }
  fmt.Printf("source pubkey: %s\n", hex.EncodeToString(tx.SourcePublicKey[:]))

  unsignedDigest, err := tx.GetUnsignedDigest()
  if err != nil {
    log.Fatalf("got err: %s when getting unsigned digest local", err.Error())
  }

  subSeed, err :=types.GetSubSeed("seed-here")
  if err!= nil {
    log.Fatalf("got err %s when getting subSeed", err.Error())
  }

  sig, err := schnorrq.Sign(subSeed, tx.SourcePublicKey, unsignedDigest)
  if err != nil {
    log.Fatalf("got err: %s when signing", err.Error())
  }
  fmt.Printf("sig: %s\n", hex.EncodeToString(sig[:]))
  tx.Signature = sig

  encodedTx, err := tx.EncodeToBase64()
  if err != nil {
    log.Fatalf("got err: %s when encoding tx to base 64", err.Error())
  }
  fmt.Printf("encodedTx: %s\n", encodedTx)

  id, err := tx.ID()
  if err != nil {
    log.Fatalf("got err: %s when getting tx id", err.Error())
  }

  fmt.Printf("tx id(hash): %s\n", id)

  url := baseUrl + "/broadcast-transaction"
  payload := struct {
    EncodedTransaction string `json:"encodedTransaction"`
  }{
    EncodedTransaction: encodedTx,
  }

  buff := new(bytes.Buffer)
  err = json.NewEncoder(buff).Encode(payload)
  if err != nil {
    log.Fatalf("got err: %s when encoding sendManyPayload", err.Error())
  }

  req, err := http.NewRequest(http.MethodPost, url, buff)
  if err != nil {
    log.Fatalf("got err: %s when creating request", err.Error())
  }

  res, err := http.DefaultClient.Do(req)
  if err != nil {
    log.Fatalf("got err: %s when performing request", err.Error())
  }
  defer res.Body.Close()

  if res.StatusCode != http.StatusOK {
    log.Fatalf("Got non 200 status code: %d", res.StatusCode)
  }

  type response struct {
    PeersBroadcasted   uint32 `json:"peersBroadcasted"`
    EncodedTransaction string `json:"encodedTransaction"`
  }
  var body response

  err = json.NewDecoder(res.Body).Decode(&body)
  if err != nil {
    log.Fatalf("got err: %s when decoding body", err.Error())
  }

  fmt.Printf("%+v\n", body)
}

```
## Error Handling
In case you receive a http resonse `400` for the endpoints\
`/ticks/{tickNumber}/approved-transactions`\
`/ticks/{tickNumber}/transactions`\
`/ticks/{tickNumber}/tick-data`\
`/ticks/{tickNumber}/quorum-data`

Make sure to implement error handling for the following cases (generally we are following the status codes from https://grpc.io/docs/guides/status-codes/):
| code   	|  reason  	| action |
|---	|---	|--- |
| 9  	|  Requested tick number is greater than `lastProcessedTick`.	| Repeat your request until it works. You may track the the `lastProcessedTick` from the endpoint `/status` until `lastProcessedTick` > your `targetTick`.  |
| 11  	|  Provided tick number was skipped by the system, next available tick is `<NEXTAVAILABLETICKNUMBER>`. 	| Take the `nextTickNumber` from `details` in response and proceed with this tick. In this case, pending transactions with `targetTick` < `nextTickNumber` won't be processed and need to be reissued.  |


## Logging
It is important to log any transaction information also on the side of the integrator. The qubic network does not store transaction information over a period longer than seven days and transactions are pruned on every epoch change each Wednesday. The RPC infrastructure provides an archive which holds historical data.

