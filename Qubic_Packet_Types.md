# Qubic Packet Types (Request/Response Types)

The following list describes the Request/Response Types of the Qubic TCP API. The number is the type which can be used in `RequestResponseHeader`.

## `EXCHANGE_PUBLIC_PEERS` 0
This is used to exchange public peers within the nodes.

## `BROADCAST_MESSAGE` 1
A General Message type used to send/receive messages from/to peers.

## `BROADCAST_COMPUTORS` 2
The List of valid Computors which are currently active.

## `BROADCAST_TICK` 3
A Tick in Qubic is one vote from a Computor.

## `BROADCAST_FUTURE_TICK_DATA` 8
This Packet contains the information from the Tickleader about a future Tick. This is the definition of the tick.

## `REQUEST_COMPUTORS` 11
This Packet can be used to ask any Peer for the current Computors list. The answer will be the `BROADCAST_COMPUTORS`.

## `REQUEST_QUORUM_TICK` 14
Request a Peer for the votes of the Computors for a specific Tick. Answer will be a bulk response of `BROADCAST_TICK`.

## `BROADCAST_TRANSACTION` 24
Packet which Contains a Qubic Transaction. A Transaction can be a Transfer (with or without flow of Qubics). A Command for a SC or any other information that will be written to the Blockchain.

## `REQUEST_CURRENT_TICK_INFO` 27
Packet to request curent state of the node. The node will reply with `RESPOND_CURRENT_TICK_INFO` which contains the current tick and state of tick votes.

## `RESPOND_CURRENT_TICK_INFO` 28
Response to `REQUEST_CURRENT_TICK_INFO`.

## `REQUEST_TICK_TRANSACTIONS` 29
Can be used to query all or specific transaction for a specific tick. The node will answer with a bulk response of `BROADCAST_TRANSACTION`.

## `REQUEST_ENTITY` 31
Can be used to query a node to get information about an Entity (Address). Node will response with `RESPOND_ENTITY`

## `RESPOND_ENTITY` 32
It includes, amoung other things, the Balance, amount of in-/outgoing transactions. Can be requested by `REQUEST_ENTITY`.

## `REQUEST_CONTRACT_IPO` 33
Can be used to query a node about Contract IPO state. It will response with `RESPOND_CONTRACT_IPO`

## `RESPOND_CONTRACT_IPO` 34
Information about IPO state of a Contract.

## `END_RESPONSE` 35
Used as a terminator for a request which will response with multiple packets.

## `REQUEST_ISSUED_ASSETS` 36
Used to query a node for issued Assets of a specific Qubic Address.

## `RESPOND_ISSUED_ASSETS` 37
Information of issued assets for a Qubic Address.

## `REQUEST_OWNED_ASSETS` 38
Used to ask a node for assets that are owned by the specified qubic address.

## `RESPOND_OWNED_ASSETS` 39
The assets owned by the address used in `REQUEST_OWNED_ASSETS`.

## `REQUEST_POSSESSED_ASSETS` 40
Used to ask a node for assets that are possesed by the specified qubic address.

## `RESPOND_POSSESSED_ASSETS` 40
The assets possesed by the address used in `REQUEST_OWNED_ASSETS`.

## `PROCESS_SPECIAL_COMMAND` 255
Used to communicate with the node for special commands.

```c++
// shut down a node (like pressing ESC)
#define SPECIAL_COMMAND_SHUT_DOWN 0ULL
// ask a node for information about current proposals
#define SPECIAL_COMMAND_GET_PROPOSAL_AND_BALLOT_REQUEST 1ULL
// response of a node about current proposals
#define SPECIAL_COMMAND_GET_PROPOSAL_AND_BALLOT_RESPONSE 2ULL
// ask a node to set a proposal
#define SPECIAL_COMMAND_SET_PROPOSAL_AND_BALLOT_REQUEST 3ULL
// response of the set command
#define SPECIAL_COMMAND_SET_PROPOSAL_AND_BALLOT_RESPONSE 4ULL
```

