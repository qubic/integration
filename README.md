# Qubic Documentation/Integration
This Documentation is meant to describe how you can access the Qubic Network API.

- [Qubic Documentation/Integration](#qubic-documentationintegration)
  - [General Concepts](#general-concepts)
  - [Basic Communication Structure](#basic-communication-structure)
  - [Examples / Use Cases](#examples--use-cases)
    - [Basics](#basics)
    - [Example Use Cases](#example-use-cases)

> [!IMPORTANT]
> The Qubic Core Main-Net runs on TCP Sockets with Port **21841**

> The Qubic Test-net runs on TCP Port **21843**

> The default Buffer size is defined in `BUFFER_SIZE` with **33554432**

## General Concepts
todo: explain used crypto, private/public key, signing/verify

| Term  | Description                                                                                     |
| ----- | ----------------------------------------------------------------------------------------------- |
| Epoch | An epoch describes a period of time in which the network runs without resetting the nodes data. |
| Tick  | A Tick is a Block in Qubic. It describes one Block in the Blockchain |
| TickLeader | The [TickLeader](Glossar/TickLeader.md) is the Computor which is responsible for a certain Tick |

## Basic Communication Structure
Communication in Qubic is basically a continious TCP Stream. All starts with the `RequestResponseHeader` which contains the information about the current Packet to be transfered.

```c++
struct RequestResponseHeader
{
private:
    unsigned char _size[3];
    unsigned char _type;
    unsigned int _dejavu;
    // ... code cutted
}
```

- `_size` => The size of the Packet. Total size including Header size.
- `_type` => Type of Packet. Refer to [Qubic Types](Qubic_Packet_Types.md)
- `_dejavu` => Marker to know if a packat was relayed, new or if we already know it. It is also used in replys to mark a response as an answer to your request.

a complete Qubic TCP Packet is defined by:

1. RequestRepsonseHeader
2. Payload

## Examples / Use Cases
We provide you sme use cases to understand how to interact with the Qubic API.

When talking to the Qubic Network and you don't operate your own Computor be careful to use **only** trusted Peers.

> [!IMPORTANT]
> To get reliable information, use only trusted Peers to request data. 

### Basics
1. [Generate Qubic Addres](UseCases//GenerateAddress.md)
2. [How to Sign](UseCases/Sign.md)

### Example Use Cases
1. [TickInfo - Basic Network Info / Blockheight](UseCases/TickInfo.md)
2. [TickData - Tick Definition (Block Info by height)](UseCases/TickData.md)
3. [Tramsaction Info](UseCases/GetTransaction.md)
4. [GetEntity - get balance of an address](UseCases/GetEntity.md)
5. [Send Transaction](UseCases/SendTransaction.md)


