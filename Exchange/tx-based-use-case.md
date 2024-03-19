# TX Based Exchange integration
The following use cases are based on the Qubic RPC V1.

The following examples refer to the Qubic RPC V1 API and the [TS Library](https://github.com/qubic/ts-library)

The TS Library uses a Webassembly for all cryptographical stuff.

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
  const response = await fetch(`https://testapi.qubic.org/ticks/${currentTick}/transfer-transactions`);
  
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
      const response = await fetch(`https://testapi.qubic.org/tx-status/${}/transfer-transactions`);
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
    const response = await fetch(`https://testapi.qubic.org/tx-status/${tickNumber}`);
    
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
