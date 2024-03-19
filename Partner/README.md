# Qubic Partner Integration

This documentation describes best practises to integrate/interact with Qubic for partners.


## Exchange integration
To partner up with Qubic from a perspective of an exchange, there are two main approaches how you can integrate Qubic into your business logic.

1. TX Based
2. Balance Based

## TX Based Workflow
This is the classical way you may already know from other blockchains. You can do block scans for deposits and send transactions via RPC call to the network.

> [Read more about the TX Based Workflow](tx-based-use-case.md).

For the transaction based approach, Qubic offers a RPC Server. Please ask us for detailed information and access.

The Qubic RPC Server is built and operated with the following software:

- `qubic-http` https://github.com/qubic/qubic-http
- `go-archiver` https://github.com/qubic/go-archiver
- `go-node-fetcher` https://github.com/qubic/go-node-fetcher
- `go-node-connector` https://github.com/qubic/go-node-connector
- `ts-library` https://github.com/qubic/ts-library


## Balance Based Workflow
In contrary to the TX Based Workflow with the balance workflow you can implement Qubics native way of integration.

In Qubic the Balance of an Address can be altered without a Transaction being issued. This can be invoked by any Smart Contract.

Therefore you can also use the Balance Based approach to interact with Qubic.

We are building the application `qubic-lrv` for this. Which is currently in testing state.

Please read here more when testing state is passed.

References:
- `qubic-lrv` https://github.com/computor-tools/qubic-lrv