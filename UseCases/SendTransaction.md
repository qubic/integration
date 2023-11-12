# Send Transaction
This Use Case describes how to send a Basic Transaction. A Bais Transaction is used to send $QBIC's from one Address to another.

To send a Transaction, send a `BroadcastTransaction` Packet with the type set to **24**.

## Request Header
- Size: **144**
- Type: **24**
- DejaVu: 0

## Request Payload
Send a `SignedTransaction`.

```c++
typedef struct
{
    unsigned char sourcePublicKey[32];
    unsigned char destinationPublicKey[32];
    long long amount;
    unsigned int tick;
    unsigned short inputType;
    unsigned short inputSize;
} Transaction;

typedef struct
{
    Transaction transaction;
    unsigned char signature[SIGNATURE_SIZE];
} SignedTransaction;
    
```

- sourcePublicKey => The Sender
- destinationPublicKey => The Receiver
- amount => Amount in $QUBIC
- tick => The Target Tick to execute the Transaction
- inputType => 0
- inputSize => 0
- signature => Signature of `Transaction` from `sourcePublicKey`

> [!NOTE]
> Set the target tick at minimum +5 from current network tick. Best results are achieved with current tick +10-20.

> [!IMPORTANT]
> Per source ID only **one** transaction can pe present in the Network. If you send another transaction with the same Source while the first was not yet executed, the existing will be overwritten.

## Response Header
none

## Response Payload
none
