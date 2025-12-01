# Migration

## Summary

This document provides guidance for partners migrating to Qubic's updated RPC infrastructure.

**Live API**: No functional changes. All services, inputs, and outputs remain the same. Only the base path has changed—endpoints are now prefixed with `/live/v1`.

**Archiver API**: Will not be available long-term. Due to scalability and architectural improvements, the Archiver API is being phased out. All endpoints are deprecated:
- **End of 2025**: First group of endpoints will be removed. Stop using immediately.
- **During 2026**: Remaining endpoints will be removed. Start migrating now to the new **Query API**.

For technical details on the architecture changes, see: [RPC 2.0 Qubic Integration Layer Functionality Upgrade](https://qubic.org/blog-detail/rpc-2-0-qubic-integration-layer-functionality-upgrade)

If you are unsure how to replace certain endpoint calls, please contact us so we can help you.

## Deprecated Archiver Endpoints

### Removed by End of 2025

The following endpoints will be removed completely by the end of 2025. Stop using these immediately:
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

### Removed During 2026

The following endpoints will be removed during 2026. Start migrating now to the Query API:
```
GET /v1/status (status service archiver status)
GET /v1/ticks/{tickNumber}/approved-transactions
GET /v1/ticks/{tickNumber}/tick-data
GET /v1/ticks/{tickNumber}/transactions
GET /v1/transactions/{txId}
GET /v1/tx-status/{txId}
GET /v2/identities/{identity}/transfers
GET /v2/ticks/{tickNumber}/transactions
GET /v2/transactions/{txId}
GET /v1/epochs/{epoch}/computors
GET /v1/latestTick
GET /v2/epochs/{epoch}/empty-ticks
GET /v2/epochs/{epoch}/ticks
```

Replacement endpoints are referenced in the [OpenAPI documentation](swagger/qubic-rpc-doc.html) under the respective deprecated entries.

## Migrating to the new Query API

To replace endpoints you can move to the new query API. This section shows how to migrate.

> Base URL: `https://rpc.qubic.org/query/v1`

Important changes compared to the old API:

- Most requests are now POST rather than GET. 
- Input must be included in the request body, not in the URL.
- More filter and range options for requests.
- Output structures have been overhauled.

> Please also refer to the openapi documentation [here](swagger/qubic-query-doc.html) and [here](swagger/qubic-rpc-doc.html) 
> and to the project documentation [here](https://github.com/qubic/archive-query-service/tree/main/v2).


### Query transaction

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

### Query tick transactions

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

### Query identity transactions

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

### Query tick data

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

### Query last processed tick

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

### Query processed tick intervals

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

### Query computors for one epoch

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
