# TickData - Tick Definition (Block Info by height)
A TickData describes one Tick (Block). It contains all important information a Comptor needs to calculate the Tick.

To query the TickData, send a `RequestedTickData` Packet with the type set to **16**.

## Request Header
- Size: **12**
- Type: **16**
- DejaVu: **any Random int**

## Request Payload
- `unsigned int tick`: The Tick for which you want to receive the TickData

## Response Header
`BROADCAST_FUTURE_TICK_DATA`

## Response Payload

```c++
typedef struct
{
    unsigned short computorIndex;
    unsigned short epoch;
    unsigned int tick;

    unsigned short millisecond;
    unsigned char second;
    unsigned char minute;
    unsigned char hour;
    unsigned char day;
    unsigned char month;
    unsigned char year;

    union
    {
        struct
        {
            unsigned char uriSize;
            unsigned char uri[255];
        } proposal;
        struct
        {
            unsigned char zero;
            unsigned char votes[(NUMBER_OF_COMPUTORS * 3 + 7) / 8];
            unsigned char quasiRandomNumber;
        } ballot;
    } varStruct;

    unsigned char timelock[32];
    unsigned char transactionDigests[NUMBER_OF_TRANSACTIONS_PER_TICK][32];
    long long contractFees[MAX_NUMBER_OF_CONTRACTS];

    unsigned char signature[SIGNATURE_SIZE];
} TickData;
```

> [!WARNING]
> A node may also return a TickData when the tick was not successful (empty). Transactions in an empty Tick will not be executed.

> todo: Describe