# Qubic partner integration

This documentation describes the best practices to integrate and interact with Qubic.

## Qubic RPC

The Qubic RPC API is the main way to interact with Qubic.

> [Swagger / OpenAPI Documentation](swagger/qubic-query-doc.html)

### API version

| Base URL               | Use case                                                                   |
|------------------------|----------------------------------------------------------------------------|
| https://api.qubic.org/ | Public API for general use-cases. Use this to build your applications.     |
| https://rpc.qubic.org/ | Deprecated legacy API. See the legacy documentation [here](old/README.md). |

> If migrating from the deprecated API, please refer to [this](migration.md) page for the endpoints that got replaced and their new counterparts.  

## Exchange integration

When integrating Qubic into your business logic, we recommend you use a TX based workflow.  
You may already be familiar with this approach from other blockchains. It mainly consists of performing block scans for deposits 
and sending transactions via an API call to the RPC infrastructure.

> [Read more about the TX based workflow](tx-based-workflow.md)
