# Qubic Exchange Integration

This documentation describes best practises to integrate/interact with Qubic from a perspective of an exchange.

There are two main approaches how you can integrate Qubic in your business logic.

1. TX Based
2. Balance Based

## TX Based
For the transaction based approach, Qubic offers a RPC Server. Please ask us for detailed information.

You can use the same tools which we also use:

- `qubic-http` https://github.com/qubic/qubic-http
- `go-archiver` https://github.com/qubic/go-archiver
- `go-node-fetcher` https://github.com/qubic/go-node-fetcher
- `go-node-connector` https://github.com/qubic/go-node-connector
- `ts-library` https://github.com/qubic/ts-library


> Please read [TX Based Use Cases](tx-based-use-case.md).

## Balance Based
In Qubic the Balance of an Address can be altered without a Transaction when the Address is involved in any Smart Contract.

Therefore you can also use the Balance Based approach to interact with Qubic.

We are building the application `qubic-lrv` for this.

https://github.com/computor-tools/qubic-lrv