# TX Based Exchange integration

> [!WARNING]
> THIS PAGE IS WORK IN PROGRESS

The following use cases are based on the Qubic RPC V1.

The following examples refer to the Qubic RPC V1 API and the [TS Library](https://github.com/qubic/ts-library)

The TS Library uses a Webassembly for all cryptographical stuff.

For all the follwing examples the `baseUrl` is set to:
```js
const baseUrl = 'https://testapi.qubic.org/v1';
```

This documentation refers to the [Qubic V1 RPC API](qubic-rpc-doc.html).

| Method  	| Endpoint    	| Description   	|  Body Payload |
|---	|---	|---	|---|
| GET  	| /block-height   	| Get the current tick/block height   	| -   |
| POST  | /broadcast-transaction	| Broadcast a transaction    	| `{ "encodedTransaction": "<BASE64RAWTX>" }  `  |
| GET  	| /ticks/{tickNumber}/approved-transactions  	| Get a List of approved transactions for the given tick 	|   - |
| GET  	| /tx-status/{txId}  	| Get the status of a single transaction 	|   - |
| GET  	| /status  	| Get the RPC status 	|   - |


## Table of Content
- [TX Based Exchange integration](#tx-based-exchange-integration)
  - [Table of Content](#table-of-content)
  - [Qubic Rules](#qubic-rules)
    - [One concurent TX per source Address](#one-concurent-tx-per-source-address)
    - [Respect RPC Status](#respect-rpc-status)
    - [Epoch change](#epoch-change)
  - [General Examples](#general-examples)
    - [Generating a Seed](#generating-a-seed)
    - [Signing a Package](#signing-a-package)
    - [Create, sign, send and verify a transaction](#create-sign-send-and-verify-a-transaction)
      - [Workflow](#workflow)
  - [Deposit Workflow](#deposit-workflow)
    - [Scan Ticks/Blocks sequentially](#scan-ticksblocks-sequentially)
      - [Special case Qutil/SendMany SC](#special-case-qutilsendmany-sc)
  - [Withdraw Workflow](#withdraw-workflow)
    - [Plain Transaction](#plain-transaction)
    - [Qutil/Send Many Smart Contract](#qutilsend-many-smart-contract)

## Qubic Rules
When using the Qubic RPC you should follow some important rules.

### One concurent TX per source Address
Due to Qubics architecture from one source address can only exist one concurrent transaction in the network.

Sample pseudo workflow:
```
Network is in tick 5.

If Alice has one tx in the network for future tick 10 and then sends another one for tick 11 it would overwrite the first one.

So, one has to wait until tick 10 has passed. 

If Alice sends a tx for tick 9, nothing would change as the TX with the higher tick set will stay in the network.
```

> [!IMPORTANT]
> Only one concurrent TX per sending ID, don’t send a new transaction from the same source address until the previous tx target tick has been passed.


### Respect RPC Status
the endpoint
- `/status`

return the current status of the RPC data.

sample status response:
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

> in the array `processedTickIntervalsPerEpoch` you find the processed ticks. it is possible due to updates, that apochs can have multiple tick ranges

### Epoch change
> WIP

## General Examples

### Generating a Seed
A seed is the private key in Qubic. Based on the seed, you can create the public address (id).

The seed is a 55-lower-case-char string. Please use a proper random generator in your environment. The Example here is for demonstration purposes only.

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

### Signing a Package
For signing a package, you can either use the qubic crypto library () or e.g. for transactions the wrapper from the `ts-library`.

the pre condition to be able to sign a package is to have the `seed` or `privateKey`.

the following example assumes that we have already created our `idPackage` which includes our `privateKey`.

```js
  // to sign a package, you need the private key which is derived from the seed and it's publicKey
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

Please find a complete example of transaction signing here: https://github.com/qubic/ts-library/blob/main/test/createTransactionTest.js The complete source code can be found in that repo too.

### Create, sign, send and verify a transaction
We assume you have already all needed data to create and send the transaction:
- Sender (including seed)
- Receiver
- Amount

#### Workflow
1. Request latest block height

```js
const response = await fetch(`${baseUrl}/block-height`);
const block = await response.json();
const latestBlockHeight = block.height;
```

2. Create and sign transaction

```js
  // please find an extended example here: https://github.com/qubic/ts-library/blob/main/test/createTransactionTest.js

  // create and sign transaction
  const tx = new QubicTransaction().setSourcePublicKey(sourcePublicKey)
      .setDestinationPublicKey(destinationPublicKey)
      .setAmount(amount)
      // it is important to set a target tick for the execution of this transaction
      // a suggested offset is 5-7 ticks
      .setTick(latestBlockHeight + 5);

    // will build the tx: bundle all values and sign it
    // returns the raw bytes of signed transaction
    const signedTransactionData = await tx.build(signSeed);

    // by requesting getId() you receive the txId of this transaction
    // this id can presented to the client as reference
    const transactionId = tx.getId();
```

3. Send transaction

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

4. Verify transaction status

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

when you ask the RPC server for `approved-transactions` you may receive a `400 Bad Request`.

sample `Bad Request` response:
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

> [!IMPORTANT]
> You must check the code to know what happened.


| code   	|  reason  	| action |
|---	|---	|--- |
| 9  	|  requested tick number `<TICKNUMBER>` is greater than last processed tick `<LASTPROCESSEDTICK>` 	| repeat your request until it works. you may track the the `LASTPROCESSEDTICK` from the endpoint `/latestTick`  |
| 11  	|  provided tick number `<TICKNUMBER>` was skipped by the system, next available tick is `<NEXTAVAILABLETICKNUMBER>` 	| take the `nextTickNumber` from `details` and proceed with this tick.  |


## Deposit Workflow
We assume that you have in your business logic the accounts of your clients. We refer to this accounts by `clientAcccount`. A client Account is a package of `seed`, `privateKey`, `publicKey` and `publicId`. The list of all `clientAccount` is called `clientAccountList`.

To detect a deposit to a `clientAccount` we use the Qubic RPC and run a sequential blockscan.
You will need to define an initial tick from which on you will start your block scans. For our example, we start with tick `13032965`.

The following code samples contains pseudo code which you have to replace by your own business logic.

### Scan Ticks/Blocks sequentially

```js
  // don't forget to do a proper errorhandling!
  // if you request a tick which is yet not processed, you will receive a 404 with a specific message

  const currentTick = 13032965;

  // request transactions for the tick
  const response = await fetch(`${baseUrl}/ticks/${currentTick}/approved-transactions`);
  
  // the result will contain all executed and approved transactions for the given tick
  const tickTransactions = await response.json();

  // map this transactions to your `clientAccountList`
  const clientDeposits = clientAccountList.filter(f => tickTransactions.find(t => t.destId == f.publicId))

  clientDeposits.forEach(clientAccount => {
    // add transaction to your accounting
    const internalTx = clientAccount.addIncomingTransactions(tickTransactions.filter(f => f.destId == clientAccount.publicId).map(m => createInternalTransaction(m)));

    // start here your internal business logic to credit your clients account
    creditClientAccount(clientAccount, internalTx);

    // start transfering the deposit to your hot wallet here
    transferToHotWallet(clientAccount, internalTx);
  });
```

repeat the above code as long you don't get a `400 Bad Request`.

#### Special case Qutil/SendMany SC
In general we suggest to not allow your clients to use their deposit accounts for smart contract usage. (e.g. pool payouts, quottery or any future use case)

But, there is a send many smart contract which you should support. A such transaction can be identified as followed:

```js
  // we assume you have in hand a tx object (e.g. from the approved-transactions endpoint)
  const tx = getTxFromApprovedTransactionsEndpoint();

  
  if(
    // address of Qutil/SendMany SC
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
      // if we are here, this tx was a sendmany sc invovation
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

        // start transfering the deposit to your hot wallet here
        transferToHotWallet(clientAccount, internalTx);
      });>

      
    }

    

```


## Withdraw Workflow
To do withdraws you can either use a plain transaction or the send many sc.
The plain transaction is limited to one transaction per hot wallet. With the send many sc you can withdraw to up to 25 clients in one transaction.

### Plain Transaction
This transaction is feeless, follow the process from [Create, sign, send and verify a transaction](#create-sign-send-and-verify-a-transaction)

### Qutil/Send Many Smart Contract
The fee for using the Smart Contract are `10` Qubic.

A send many smart contract invocation is a qubic transactions with some specific settings.

the below example shows how to use it.

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

    // after a tx has been send you can either check its status by block scan or by dedicated call to /tx-status/<TDID>

  }else {
    // :( something went wrong, try again
  }


```


