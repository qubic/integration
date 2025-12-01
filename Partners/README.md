# Qubic partner integration

This documentation describes the best practices to integrate and interact with Qubic.

## Qubic RPC

The Qubic RPC API is the main way to interact with Qubic.

For API specifications, see the 
[Swagger / OpenAPI Documentation](swagger/qubic-query-doc.html).

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

You should only use the query and live API.

## Exchange integration

When integrating Qubic into your business logic, we recommend you use a TX based workflow.  
You may already be familiar with this approach from other blockchains. It mainly consists of performing block scans for deposits 
and sending transactions via an API call to the RPC infrastructure.

> [Read more about the TX based workflow](tx-based-workflow.md)


## Migration and Legacy Documentation

The old Archiver API has been replaced by the new Query API. The Archiver API is still available for a while for backwards compatibility, but all endpoints are deprecated and will be removed.

⚠️ **The Archiver API will not be available long-term.** All endpoints are deprecated:
- **End of 2025**: First group of endpoints will be removed.
- **During 2026**: Remaining endpoints will be removed.

If you are migrating from the deprecated API, please refer to the [Migration Guide](migration.md) for the endpoints that got replaced and their new counterparts.

For technical details on the architecture changes, see: [RPC 2.0 Qubic Integration Layer Functionality Upgrade](https://qubic.org/blog-detail/rpc-2-0-qubic-integration-layer-functionality-upgrade)

For legacy documentation, see [here](old/README.md).
