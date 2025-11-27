# Qubic partner integration

This documentation describes the best practices to integrate and interact with Qubic.

## Qubic RPC

The Qubic RPC API is the main way to interact with Qubic.

[Swagger / OpenAPI Documentation](swagger/qubic-query-doc.html)

### API

| Host                   | Use case                                         |
|------------------------|--------------------------------------------------|
| https://rpc.qubic.org/ | Public API. Use this to build your applications. |

There are APIs with different purpose. For the partner integration the query and live API are relevant. For legacy
integrations the archive API was relevant.

![integration-apis](integration-apis.png)

* `Query API` ... API for querying archived data. Base path: `/query/v1`
* `Live API` ... API that queries live data from the qubic nodes. Base path: `/live/v1`
* `Archiver API` ... deprecated/legacy API for querying archived data directly from the old archiver.
* `Stats API` ... not relevant for partner integration. Integration layer for the explorer.

You should only use the query and live API. The stats api can be changed without further notice.

## Exchange integration

When integrating Qubic into your business logic, we recommend you use a TX based workflow.  
You may already be familiar with this approach from other blockchains. It mainly consists of performing block scans for deposits 
and sending transactions via an API call to the RPC infrastructure.

> [Read more about the TX based workflow](tx-based-workflow.md)

## Migration and legacy documentation

The current API is different from the original one. Mainly a new query API for accessing the archive was created and the
old archiver API is deprecated. Some paths of the old archiver API are still available for backwards compatibility. Some
old archiver endpoints are removed.

If migrating from the deprecated API, please refer to [this](migration.md) page for the endpoints that got replaced and 
their new counterparts.

If you want to access the old documentation you can find it [here](old/README.md).
