# Migration

Migration is only necessary for the old archiver endpoints. The live endpoints will stay available (but prefixed with the base path `/live`).

## Deprecated archiver endpoints

The following endpoints will be removed completely as soon as possible (ETA end of 2025):

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

For these endpoints users need to migrate ASAP.

The following endpoints will be available for a longer migration period:

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

For these endpoints immediate migration is not necessary but should be done in 2026.

## Migrating to the new Query API

All the old archiver endpoints can be replaced with the new query endpoints (if there is a gap for you please let us know).

> Base URL: `https://rpc.qubic.org/query/v1`

Important changes compared to the old API:  
- Input data is now passed in the request body.
- More filter and range options for requests.
- Output structures have been overhauled.

> Please also refer to the swagger documentation found [here](swagger/qubic-query-doc.html) and to the project documentation found [here](https://github.com/qubic/archive-query-service/tree/main/v2).


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
