# TickInfo - Basic Network Info / Blockheight
To get the current Blockheight, you can use the `REQUEST_CURRENT_TICK_INFO` request.

To query the Tick Info, send a `RequestResponseHeader` Packet with the type set to **27**.

## Request Header
- Size: **8**
- Type: **27**
- DejaVu: **any Random int**

## Request Payload
- empty

## Response Header
`RESPOND_CURRENT_TICK_INFO`

## Response Payload
you wil get the following response:

```c++
typedef struct
{
    unsigned short tickDuration;
    unsigned short epoch;
    unsigned int tick;
    unsigned short numberOfAlignedVotes;
    unsigned short numberOfMisalignedVotes;
} CurrentTickInfo;
```

All information in the response are related to the nodes status:

- tickDuration: How long the current Node stays in that tick
- epoch: the epoch
- **tick**: the tick; current Blockheight this node stays
- numberOfAlignedVotes: the votes from Computors the peer sees which are the same as he already have
- numberOfMisalignedVotes: the votes from Computors the peer sees which are different from it's own view

> [!NOTE]
> It makes sense to keep track of the Tick and save it in your application. If you get a response from a node with a lower Tick, ignore that packet. The peer seems to be not up to date.