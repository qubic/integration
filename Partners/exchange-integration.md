# TX Based Exchange integration

## Management Summary
This document describes the Qubic RPC infrastructure and how it can be used to integrate Qubic into an exchange environment.

## About qubic
* Launched: 13.4.2022 (mainnet)
* Fair Launch: no VC or team allocations
* Circulating Supply: 71.4trn (trilion)
* Weekly Emission: 1trn
* Hard Cap/Max Supply: 1000trn (will most likely never be reached due to burn mechanisms)

### Available to buy
* [SafeTrade](https://safe.trade/)
* [Tradeogre](https://tradeogre.com/markets)
* [Seven Seas Exchange](https://www.sevenseas.exchange/)
* [Bitkonan](https://www.bitkonan.com)

### Trackers
* [Coinmarketcap](https://coinmarketcap.com/currencies/qubic/)
* [Coingecko](https://www.coingecko.com/en/coins/qubic-network)
* [Delta mobile app](http://delta.app)
* [Livecoinwatch](https://www.livecoinwatch.com/price/QUBIC-QUBIC)
* [Coinpaprika](https://coinpaprika.com/coin/qubic-qubic/)


## Exchange Related Information
* Project Name: Qubic
* Ticker: QUBIC 
* Currency: QU
* Brand assets (logo, icon) for [download](https://drive.google.com/file/d/13XcdR7nWMTNVRL5J5FYZyc-SnMMAYN3w/view)
* Decimals: The smallest possible unit is 1 QU (like 1 Satoshi in Bitcoin) and it trades at a price of around 0.000006879 USD right now.

### Links
* Github: [https://github.com/qubic](https://github.com/qubic)
* X/Twitter: [https://x.com/_Qubic_](https://x.com/_Qubic_) 
* Website: [https://qubic.org/](https://qubic.org/)
* Discord: [https://discord.com/invite/qubic](https://discord.com/invite/qubic)
* Telegram: [https://t.me/qubic_network](https://t.me/qubic_network)
* Youtube: [https://www.youtube.com/@_qubic_](https://www.youtube.com/@_qubic_)
* Explorer: [https://app.qubic.li/network/explorer](https://app.qubic.li/network/explorer) 
* Medium, Reddit, Facebook, Instagram: n/a

### Tech Integration Basics
* Base URL for production environment: `https://ex.qubic.org/v1`
* Base URL for test environment (points to test infrastructure which accesses the production chain): `https://testapi.qubic.org/v1`
* Access to the production environment is IP restricted.
* Individual rate limits can be set by source ip address.
* The HTTP API is open and doesn’t require any special authentication.
* All infrastructure is protected by Cloudflare

> [!IMPORTANT]
> Use `https://testapi.qubic.org/v1` for testing and let us know your IP addresses in order to access `https://ex.qubic.org/v1`

### Tech FAQ
* **Is there a need to run an own node?** \
No, we suggest using our infrastructure as described in this document for easy onboarding. But you can also run your own node. Let’s discuss.
* **Does qubic support memo for wallets?** \
No, it doesn’t.
* **Where do I get test tokens?** \
Just get in touch, we’re happy to send some. \


# General Qubic Terms
This section is about terms needed to understand Qubic. It explains the dos and don’ts.

## QU (Qubics)
QU is the native currency of the Qubic network.

## Epoch
An Epoch in Qubic is one week. It starts every Wednesday at 12 UTC and ends one week later before the new epoch starts.
Epochs are important because every epoch change gives the node operators the chance to update their nodes. Currently every week is expected to have one core update.
During epoch change ~11:00 to ~13:00 UTC the general network availability is deprecated. This means that asking for current balance may take a bit longer than normal.
Using the Qubic RPC interface you will have responsiveness also during epoch transition because all important data is cached.

> [!IMPORTANT]
> We suggest <strong>not</strong> to  send out transactions during the transition phase. Every Wednesday 11:45 UTC and 12:45 UTC.

## Computor
A Computor is one of 676 validators in Qubic.
Computors take on a multitude of tasks, among them are validation of ticks, execution of smart contracts, and validation of transactions. 

## Quorum
The Quorum (derived from a paper of [Nick Szabo](https://www.fon.hum.uva.nl/rob/Courses/InformationInSpeech/CDROM/Literature/LOTwinterschool2006/szabo.best.vwh.net/quorum.html), see also: [Quorum in distributed systems on Wikipedia](https://en.wikipedia.org/wiki/Quorum_(distributed_computing)). ) in Qubic is defined by 451 of 676 Computors.
The Quorum is needed to finalize Ticks or to make decisions. Only if 451 out of 676 Computors agree to a common state the network can proceed.
The Quorum is also needed to make network related decisions. This can be achieved by proposals and votes.

## Tick
A Tick in Qubic is a block for other blockchains. A Tick is defined by its TickData which is issued by the responsible Computor. A tick usually takes between two and eight seconds.
Every Tick can be non-empty or empty. Only non-empty ticks are considered as good and will include and execute transactions.
As soon as a tick has passed, the transactions (and everything else) included cannot be changed or reverted anymore (true finality).

> [!IMPORTANT]
> If a Tick is empty, transactions from this Tick are not included into blockchain and therefore not executed.

> [!IMPORTANT]
> Working with Ticks is the most important concept to understand in qubic. Integration will need to monitor ticks and act according to outcome (ie. re-issue a transaction or issue a new one - more on this later in this document)

## TickData
TickData is the definition of a Tick. It contains the cryptographically signed information what transactions must be processed in this Tick.
It also contains the timestamp of the Tick and other block information needed to determine if a Tick is valid or not.

## Transaction
A transaction in Qubic is a transport vehicle for multiple purposes.
The main purposes are:
1. Sending QU from one address to another
2. Sending Smart Contract commands

A transaction in Qubic needs to have a `Target Tick` when it should be executed. The network cannot decide this. It is good practice to add 5 Ticks of offset to the `latest Tick` (current block height).

> [!IMPORTANT]
>A Qubic transaction must have a `Target Tick`. The `Target Tick` can be chosen from `latest Tick` + 5

Due to Qubics architecture from one source address can **only exist one concurrent transaction** in the network.
Network is in tick 5. If Alice has one tx in the network for future tick 10 and then sends another one for tick 11 it would overwrite the first one. So one has to wait until tick 10 has passed. If Alice sends a tx for tick 9, nothing would change as the TX with the higher tick set will stay in the network.

> [!IMPORTANT]
> Only one concurrent TX per sending ID can exist. Don’t send a new transaction from the same source address until the tx `Target Tick` has been passed. This is important to consider for withdrawals: a hot wallet can only have one transaction, everything else needs to be queued. We discuss an approach ("SendMany") which allows for more bandwidth further down.

Transactions can be created and signed locally/offline. Only the signed transaction must be sent to the network.

# The Integration for Exchanges
The following documentation showcases the integration of qubic for exchanges: deposits and withdrawals based on the RPC services provided.

Code examples leverage the [TS Library](https://github.com/qubic/ts-library)

Integrators may also use:
* [Go Library](https://github.com/)
* [HTTP Library](https://github.com/)

> [!WARNING]
> Library links missing

The TS Library uses a Webassembly for all cryptographical stuff.

For all the follwing examples the `baseUrl` is set to:
```js
const baseUrl = 'https://testapi.qubic.org/v1';
```

This documentation refers to the Qubic V1 RPC API.

| Method  	| Endpoint    	| Description   	| 
|---	|---	|---	|---
| GET  	|   	|   	|   
|   	|   	|   	|   
|   	|   	|   	|   

> [!WARNING]
> Links missing

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

### Error Handling
If any request has a problem, you will get a proper HTTP status code.

<table>
  <tr>
   <td>Code
   </td>
   <td>Description
   </td>
  </tr>
  <tr>
   <td>200
   </td>
   <td>OK, all good
   </td>
  </tr>
  <tr>
   <td>4xx
   </td>
   <td>Bad Request: You have any error in your request
   </td>
  </tr>
  <tr>
   <td>5xx
   </td>
   <td>Internal Server Error: there is an internal problem
   </td>
  </tr>
</table>

If you receive a 4xx response it is most probably because you sent anything which is not expected. In the response body you will find a description.
If you receive a 5xx response anything went wrong on our side. If you keep getting those responses, please contact us.
Generally, we apply the Status Code Handling from GRPC Core ([https://grpc.github.io/grpc/core/md_doc_statuscodes.html](https://grpc.github.io/grpc/core/md_doc_statuscodes.html)) 

### Respect RPC Status
the endpoint
- `/status`

returns the current status of the RPC data.

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

#### Full Workflow to Issue Transactions
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
| X  	|   	| |


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

