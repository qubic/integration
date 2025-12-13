# RPC APIs Migration Guide

## Table of Contents

- [Summary](#summary)
- [Live API Updates](#live-api-updates)
- [Archiver API Deprecation](#archiver-api-deprecation)
  - [Endpoints Available Until January 2026](#endpoints-available-until-january-2026)
  - [Endpoints with Query API Replacements](#endpoints-with-query-api-replacements)
- [Query API Migration](#query-api-migration)
  - [Endpoint Migration Summary](#endpoint-migration-summary)
  - [Migration Examples](#migration-examples)
- [Contact](#contact)

## Summary

This document provides guidance for partners migrating to Qubic's updated RPC infrastructure, redesigned for improved scalability and more flexible APIs. For technical details, see: [RPC 2.0 Qubic Integration Layer Functionality Upgrade](https://qubic.org/blog-detail/rpc-2-0-qubic-integration-layer-functionality-upgrade).

Here's what changed for each API:

**Live API**: No functional changes—only the base path has changed. See [Live API Updates](#live-api-updates).

**Archiver API**: Being phased out. See [Archiver API Deprecation](#archiver-api-deprecation) for details and timelines.

**Query API**: New API replacing the Archiver API for accessing archived tickchain data. See [Query API Migration](#query-api-migration).

## Live API Updates

No functional changes. All services, inputs, and outputs remain the same. The only change is that `/live` has been added to the path before `/v1`.

> Base URL: `https://rpc.qubic.org/live/v1`

> [!NOTE]
> The old paths (without `/live`) will remain supported until end of 2026. However, we advise updating your systems as soon as possible.

For example, `/v1/tick-info` becomes `/live/v1/tick-info`.

For full API specifications, see the [Live API OpenAPI documentation](https://qubic.github.io/integration/Partners/swagger/qubic-rpc-doc.html?urls.primaryName=Qubic%20RPC%20Live%20Tree).

## Archiver API Deprecation

The Archiver API is being phased out and will be **available until end of 2026**. Below are the affected endpoints grouped by timeline and action required.

### Endpoints Available Until January 2026

> [!CAUTION]
> **Action required:** These endpoints will be removed from the Archiver API by end of January 2026. Stop using them immediately.


```
GET /v1/healthcheck
GET /v1/identities/{identity}/transfer-transactions
GET /v1/ticks/{tickNumber}/chain-hash
GET /v1/ticks/{tickNumber}/quorum-tick-data
GET /v1/ticks/{tickNumber}/store-hash
GET /v1/ticks/{tickNumber}/transfer-transactions
GET /v2/ticks/{tickNumber}/hash
GET /v2/ticks/{tickNumber}/quorum-data
GET /v2/ticks/{tickNumber}/store-hash
GET /v2/transactions/{txId}/sendmany
```

> **Note:** As a temporary workaround, some v1 endpoints have v2 equivalents (available until Arhiver API deprecation):
> - `GET /v1/identities/{identity}/transfer-transactions` → `GET /v2/identities/{identity}/transfers`
> - `GET /v1/ticks/{tickNumber}/transfer-transactions` → `GET /v2/ticks/{tickNumber}/transactions?transfers=true`

### Endpoints with Query API Replacements

> [!WARNING]
> These endpoints have replacements in the Query API. Migrate during 2026 before the Archiver API is phased out.

```
GET /v1/epochs/{epoch}/computors
GET /v1/latestTick
GET /v1/status (status service archiver status)
GET /v1/ticks/{tickNumber}/approved-transactions
GET /v1/ticks/{tickNumber}/tick-data
GET /v1/ticks/{tickNumber}/transactions
GET /v1/transactions/{txId}
GET /v1/tx-status/{txId}
GET /v2/epochs/{epoch}/empty-ticks
GET /v2/epochs/{epoch}/ticks
GET /v2/identities/{identity}/transfers
GET /v2/ticks/{tickNumber}/transactions
GET /v2/transactions/{txId}
```

## Query API Migration

This section shows how to migrate from the deprecated Archiver API to the new Query API. For implementation details, see the [Archive Query Service repository](https://github.com/qubic/archive-query-service/tree/main/v2).

> Base URL: `https://rpc.qubic.org/query/v1`

Important changes compared to the old API:

- Most requests are now POST rather than GET.
- Input must be included in the request body, not in the URL.
- More filter and range options for requests.
- Output structures have been overhauled.

### Endpoint Migration Summary

| Old Archiver API                                      | New Query API                                                                                               |
| ----------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- |
| `GET /v1/transactions/{txId}`                         | `POST /query/v1/getTransactionByHash`                                                                       |
| `GET /v1/tx-status/{txId}`                            | `POST /query/v1/getTransactionByHash`                                                                       |
| `GET /v2/transactions/{txId}`                         | `POST /query/v1/getTransactionByHash`                                                                       |
| `GET /v1/ticks/{tickNumber}/transactions`             | `POST /query/v1/getTransactionsForTick`                                                                     |
| `GET /v1/ticks/{tickNumber}/approved-transactions`    | `POST /query/v1/getTransactionsForTick`                                                                     |
| `GET /v1/ticks/{tickNumber}/transfer-transactions`    | `POST /query/v1/getTransactionsForTick`                                                                     |
| `GET /v2/ticks/{tickNumber}/transactions`             | `POST /query/v1/getTransactionsForTick`                                                                     |
| `GET /v1/identities/{identity}/transfer-transactions` | `POST /query/v1/getTransactionsForIdentity`                                                                 |
| `GET /v2/identities/{identity}/transfers`             | `POST /query/v1/getTransactionsForIdentity`                                                                 |
| `GET /v1/ticks/{tickNumber}/tick-data`                | `POST /query/v1/getTickData`                                                                                |
| `GET /v1/status`                                      | `GET /query/v1/getLastProcessedTick`                                                                        |
| `GET /v1/status`                                      | `GET /query/v1/getProcessedTickIntervals`                                                                   |
| `GET /v1/latestTick`                                  | `GET /live/v1/tick-info` (latest network tick) or `GET /query/v1/getLastProcessedTick` (last archived tick) |
| `GET /v1/epochs/{epoch}/computors`                    | `POST /query/v1/getComputorListsForEpoch`                                                                   |

> [!NOTE]
> For full API specifications, see the [Query OpenAPI documentation](swagger/qubic-query-doc.html).

### Migration Examples

#### Query transaction

Old  
```HTTP
GET /v2/transactions/stvdxfctjgsqvcvnloqrqsyikligbofdqvzmoqqficfarnzpuknxxqrectbj HTTP/1.1
Host: rpc.qubic.org
Accept: application/json
```

New
```HTTP
POST /query/v1/getTransactionByHash HTTP/1.1
Host: rpc.qubic.org
Content-Type: application/json
Accept: application/json

{
  "hash": "stvdxfctjgsqvcvnloqrqsyikligbofdqvzmoqqficfarnzpuknxxqrectbj"
}
```

#### Query tick transactions

Old
```HTTP
GET /v2/ticks/18997135/transactions?transfers=false&approved=false HTTP/1.1
Host: rpc.qubic.org
Accept: application/json
```

New
```HTTP
POST /query/v1/getTransactionsForTick HTTP/1.1
Host: rpc.qubic.org
Content-Type: application/json
Accept: application/json

{
  "tickNumber": 18997135
}
```

#### Query identity transactions

Old
```HTTP
GET /v2/identities/DVBCVWTPOFUGCFEPMJJEEHQKQYCAANJFXYXUNVQWPEFQOCBINNCGGVFCJVYJ/transfers?startTick=15123305&endTick=18997135&scOnly=false&desc=true&page=1&pageSize=10 HTTP/1.1
Host: rpc.qubic.org
Accept: application/json
```

New
```HTTP
POST /query/v1/getTransactionsForIdentity HTTP/1.1
Host: rpc.qubic.org
Content-Type: application/json
Accept: application/json

{
  "identity": "DVBCVWTPOFUGCFEPMJJEEHQKQYCAANJFXYXUNVQWPEFQOCBINNCGGVFCJVYJ",
  "filters": {
  },
  "ranges": {
    "tickNumber": {
      "gte": "15123305",
      "lte": "18997135"
    },
    "amount": {
        "gt": "0"
    }
  },
  "pagination": {
    "offset": 0,
    "size": 10
  }
}
```

> See https://github.com/qubic/archive-query-service/tree/main/v2 for more details about filters and ranges.

#### Query tick data

Old
```HTTP
GET /v1/ticks/18997135/tick-data HTTP/1.1
Host: rpc.qubic.org
Accept: application/json
```

New
```HTTP
POST /query/v1/getTickData HTTP/1.1
Host: rpc.qubic.org
Content-Type: application/json
Accept: application/json

{
  "tickNumber": 18997135
}
```

#### Query last processed tick

Old
```HTTP
GET /v1/status HTTP/1.1
Host: rpc.qubic.org
Accept: application/json
```

New
```HTTP
GET /query/v1/getLastProcessedTick HTTP/1.1
Host: rpc.qubic.org
Accept: application/json
```

#### Query processed tick intervals

Old
```HTTP
GET /v1/status HTTP/1.1
Host: rpc.qubic.org
Accept: application/json
```

New
```HTTP
GET /query/v1/getProcessedTickIntervals HTTP/1.1
Host: rpc.qubic.org
Accept: application/json
```

#### Query computors for one epoch

Old
```HTTP
GET /v1/epochs/179/computors HTTP/1.1
Host: rpc.qubic.org
Accept: application/json
```

New
```HTTP
POST /query/v1/getComputorListsForEpoch HTTP/1.1
Host: rpc.qubic.org
Content-Type: application/json
Accept: application/json

{
  "epoch": 179
}
```

## Contact

Have questions about the migration? Reach out to us on [Discord](https://discord.com/channels/768887649540243497/1087017597133922474).
