# Qubic partner integration

This documentation describes the best practices to integrate and interact with Qubic.

## Qubic RPC

The Qubic RPC API is the main way to interact with Qubic.

| Host                   | Use case                                          |
|------------------------|---------------------------------------------------|
| https://rpc.qubic.org  | Public API. Use this to build your applications.  |

### Available APIs

For partner integration, use the **Query API** and **Live API** only. The Archiver API is deprecated and will be removed.

![integration-apis](integration-apis.png)

| API           | Purpose                                 | Base Path    | Status               |
|---------------|-----------------------------------------|--------------|----------------------|
| **Query API** | Querying archived data                  | `/query/v1`  | ✅ Active             |
| **Live API**  | Querying live data from Qubic nodes     | `/live/v1`   | ✅ Active             |
| Archiver API  | Legacy API for archived data            | `/v1`, `/v2` | ⚠️ Deprecated        |
| Stats API     | Market data, rich lists, and statistics | —            | 🔄 Subject to change |

> **Note:** The Stats API is not intended for partner integration. It is currently under evaluation and may change without notice.

For API specifications, see the [Openapi Documentation](https://qubic.github.io/integration/Partners/swagger/qubic-rpc-doc.html).

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
