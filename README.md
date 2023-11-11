# Qubic Integration
This Documentation is meant to describe how you can access the Qubic Network API.

> [!IMPORTANT]
> The Qubic Core Main-Net runs on TCP Sockets with Port **21841**

> The Qubic Test-net runs on TCP Port **21843**

> The default Buffer size is defined in `BUFFER_SIZE` with **33554432**

## General Concepts
todo: explain used crypto, private/public key, signing/verify

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



