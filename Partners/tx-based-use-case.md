# TX Based Exchange integration

> [!WARNING]
> THIS PAGE IS WORK IN PROGRESS

The following use cases are based on the Qubic RPC V1.

The following examples refer to the Qubic RPC V1 API and the [TS Library](https://github.com/qubic/ts-library)

The TS Library uses a Webassembly for all cryptographical stuff.

For all the follwing examples the `baseUrl` is set to:
```js
const baseUrl = 'https://testapi.qubic.org';
```


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
    - [1. Scan Ticks/Blocks sequentially](#1-scan-ticksblocks-sequentially)
    - [2. Verify incoming transactions](#2-verify-incoming-transactions)

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
Both endpoints:
- `/status`
- `/tx-status/status`

return the current status of the RPC data.

sample status response:
```json
{
    "lastProcessedTick": 13054803,
    "lastProcessedTicksPerEpoch": {
        "99": "12941678",
        "100": "13054803"
    },
    "skippedTicks": [
        {
            "startTick": 12941679,
            "endTick": 12949999
        }
    ]
}
```

> [!IMPORTANT]
> The `lastProcessedTick` indicates to which tick the RPC server has processed data.
> All endpoints below `/tx-status` are related to `/tx-status/status`
> All other endpoints are related to `/status`

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
const response = await fetch(`${baseUrl}/live/block-height`);
const block = await response.json();
const latestBlockHeight = block.height;
```

1. Create and sign transaction

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

1. Send transaction

```js
  // after creating and signing the tx it should be sent to the network
  const response = fetch(`${baseUrl}/live/broadcast-transaction`,
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

1. Verify transaction status

```js
  // you can verify if a transaction was successful as soon the target tick has passed (true finality)

  // make sure, that the /tx-status/status => latestProcessedTick is >= target tick of tx

   const response = await fetch(`${baseUrl}/tx-status/ticks/${tx.tick}`);
      // only if status == 200 (ok) the tx can be marked as verified
    if(response.status == 200) 
    {
      const tickTransactions = await response.json();
      const txStatus = tickTransactions.transactionsStatus.find(f => f.txId === tx.getId());
      if(txStatus && txStatus.moneyFlew){
        // yipiii! transaction was successful
      }else {
        // :( transaction was NOT successful
      }
    }else if(response.status === 404){
      const errorStatus = await response.json();
      if(errorStatus.code === 123){
        // your request was to early, please repeat it
        // the lastProcessedTick is lower than the one you have requested
      }
    }

```

## Deposit Workflow
We assume that you have in your business logic the accounts of your clients. We refer to on of accounts this by `clientAcccount`. A client Account is a package of `seed`, `privateKey`, `publicKey` and `publicId`. The list of all `clientAccount` is called `clientAccountList`.

To detect a deposit to a `clientAccount` we use the Qubic RPC and run a sequential blockscan.
You will need to define an initial tick from which on you will start your block scans. For our example, we start with tick `13032965`.

The following code samples contains pseudo code which you have to replace by your own business logic.

### 1. Scan Ticks/Blocks sequentially

```js
  // don't forget to do a proper errorhandling!
  // if you request a tick which is yet not processed, you will receive a 404 with a specific message

  const currentTick = 13032965;

  // request transactions for the tick
  const response = await fetch(`${baseUrl}/ticks/${currentTick}/transfer-transactions`);
  
  // the result will contain all executed transactions for the given tick
  const tickTransactions = await response.json();

  // map this transactions to your `clientAccountList`
  const clientDeposits = clientAccountList.filter(f => tickTransactions.find(t => t.destId == f.publicId))

  clientDeposits.forEach(clientAccount => {
    clientAccount.addIncomingTransactions(tickTransactions.filter(f => f.destId == clientAccount.publicId).map(m => createInternalTransaction(f)));
  });
```

repeat the above code as long you don't get a 404 and increase `currentTick` for each iteration.

### 2. Verify incoming transactions
After you identified potential incoming transfers for your clients you need to verify them.

You can either verify every single transactions:
```js
  // don't forget to to a proper errorhandling!

  clientAccountList.filter(f => f.hasUnverifiedTransaction).map(clientAccount =>
  {
    clientAccount.transactions.filter(f => !f.isVerified).map(unverifiedTransaction => {
      const response = await fetch(`${baseUrl}/tx-status/${}/transfer-transactions`);
      // only if status == 200 (ok) the tx can be marked as verified
      if(response.status == 200) 
      {
        const txStatus = await response.json();
        if(txStatus.txId === unverifiedTransaction.txId){
          unverifiedTransaction.isVerified = true;
          // the moneyFlew indicates if on behalf of the executed transaction qubics are flew
          unverifiedTransaction.moneyFlew = txStatus.moneyFlew;

          // start here your internal business logic to credit your clients account
          if(unverifiedTransaction.moneyFlew) {
            creditClientAccount(clientAccount, unverifiedTransaction);

            // start transfering the deposit to your hot wallet here
            transferToHotWallet(clientAccount, unverifiedTransaction);
          }

        }
      }
    });

  });
```

or, you can also do tick/block verification:
```js
  // don't forget to to a proper errorhandling!
  // if you request a tick which is yet not processed, you will receive a 404 with a specific message

  async function blockTxVerify(tickNumber){
    // request transactions for the tick
    const response = await fetch(`${baseUrl}/tx-status/${tickNumber}`);
    
    if(reponse.status !== 200)
    {
      // something went wrong; repeat this later again
      return;
    }

    // the result will contain the tx status for each tx in the given tick
    const txStatuses = await response.json();
    
    // load alread stored transfers for the tick
    const allClientTransfers = getAllClientTransferForTick(tickNumber);

    allClientTransfers.forEach(transfer => {
      const txStatus = txStatuses.find(f => f.txId = transfer.txId);
      if(txStatus){
        transer.isVerified = true;
        transfer.moneyFlew = txStatus.moneyFlew;

         // start here your internal business logic to credit your clients account
          if(transfer.moneyFlew) {
            creditClientAccountByTransaction(transfer);

            // start transfering the deposit to your hot wallet here
            transferToHotWallet(transfer);
          }

      }

    }):
    
  }

```
