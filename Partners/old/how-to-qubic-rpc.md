# How to build your own Qubic RPC Server?
In a trustless world it could make sense to operate your own RPC server. Qubic is a decentralized network. All our software is open source, also the Qubic RPC server.

> [!IMPORTANT]
> THIS DOC IST WORK IN PROGRESSS


## Pre requisites
To operate your own RPC Server you will need:

1. Computornodes under your control.
2. A Dedicated or Cloud Server with Linux (Ubuntu 22.04)
3. Good internet connection for 1) and 2)
4. Some time and patience to 1. build the RPC Server and 2. to operate it
5. Good technical knowledge to operate servers (including docker, linux configurations and operating Qubic nodes)
6. Basic knowledge of coding

## What do you get?
When you build your own RPC server you will get a full Qubic archive. This archive can be used to automate your business with Qubic.

In the time of writing this doc (April, 2024), the full Qubic archive includes (always from the point in time when you start your server):

- Archive of all Transactions (including their status)
- Persisting of all relevant Epoch information
  - Computors
  - Tick's and Tickdata
  - Quorum votes

## What software is needed to build the Qubic RPC server?
The Qubic RPC server is built on the following software packages:

- `qubic-http` https://github.com/qubic/qubic-http
- `go-archiver` https://github.com/qubic/go-archiver
- `go-node-fetcher` https://github.com/qubic/go-node-fetcher
- `go-node-connector` https://github.com/qubic/go-node-connector
- `tx-status-request add-on` https://github.com/qubic/core/tree/feature/tx-status-request
- `traefik` https://github.com/traefik/traefik

## Setup instructions
coming soon
  