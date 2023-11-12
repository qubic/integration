# Title
Basic Descriptiong

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

```

> [!WARNING]
> A node may also return a TickData when the tick was not successful (empty). Transactions in an empty Tick will not be executed.
