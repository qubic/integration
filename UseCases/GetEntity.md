# GetEntity - get balance of an address
To query the Balance of a specific Address you can use the `REQUEST_ENTITY` with type **31**

## Request Header
- Size: **40**
- Type: **31**
- DejaVu: **any Random int**

## Request Payload
- `unsigned char publicKey[32];`: PublicKey of the Address you want to get balance

## Response Header
`RESPOND_ENTITY`

## Response Payload

```c++
struct Entity
{
    unsigned char publicKey[32];
    long long incomingAmount, outgoingAmount;
    unsigned int numberOfIncomingTransfers, numberOfOutgoingTransfers;
    unsigned int latestIncomingTransferTick, latestOutgoingTransferTick;
};
typedef struct
{
    ::Entity entity;
    unsigned int tick;
    int spectrumIndex;
    unsigned char siblings[SPECTRUM_DEPTH][32];
} RespondedEntity;
```

- `publicKey`: Address/Id
- `incomingAmount`: Amount of incoming $QUBIC's for given Address
- `outgoingAmount`: Amount of incoming $QUBIC's for given Address
- `numberOfIncomingTransfers`: Amount of incoming Transactions for given Address
- `numberOfOutgoingTransfers`: Amount of incoming Transactions for given Address
- `latestIncomingTransferTick`: Tick for the last incoming Transaction
- `latestOutgoingTransferTick`: Tick for the last outgoing Transaction
- `tick`: For which Tick this Report is valid
- `spectrumIndex`: Position of Entity in Spectrum

> [!NOTE]
> To get current Balance for the given Address substract `outgoingAmount` from `incomingAmount`

> [!NOTE]
> Keep track of `tick` in your application. Don't update Balance if responded tick is lower than existing.

> [!WARNING]
> A manipulated node may report wrong information. Only use trusted Peers!
