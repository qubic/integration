# Migrating to the new Query service


> Base URL: `https://api.qubic.org`

Important changes compared to the old API:  
- Input data is now passed in the request body.
- More filter and range options for requests.
- Output structures have been overhauled.

> Please also refer to the swagger documentation found [here](swagger/qubic-query-doc.html) and to the project documentation found [here](https://github.com/qubic/archive-query-service/tree/main/v2).


## Transaction related queries

### Query transaction

Old  
```HTTP
GET /v2/transactions/stvdxfctjgsqvcvnloqrqsyikligbofdqvzmoqqficfarnzpuknxxqrectbj HTTP/1.1
Host: rpc.qubic.org
Accept: application/json
```

New
```HTTP
POST /getTransactionByHash HTTP/1.1
Host: api.qubic.org
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
POST /getTransactionsForTick HTTP/1.1
Host: api.qubic.org
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
POST /getTransactionsForIdentity HTTP/1.1
Host: api.qubic.org
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

> See https://github.com/qubic/archive-query-service/tree/main/v2 for more info about filters and ranges.


## Tick related queries

### Query tick data

Old
```HTTP
GET /v1/ticks/18997135/tick-data HTTP/1.1
Host: rpc.qubic.org
Accept: application/json
```

New
```HTTP
POST /getTickData HTTP/1.1
Host: api.qubic.org
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
GET /getLastProcessedTick HTTP/1.1
Host: api.qubic.org
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
GET /getProcessedTicksIntervals HTTP/1.1
Host: api.qubic.org
Accept: application/json
```

## Epoch related queries

Old
```HTTP
GET /v1/epochs/179/computors HTTP/1.1
Host: rpc.qubic.org
Accept: application/json
```

New
```HTTP
GET /getComputorsListForEpoch?epoch=179 HTTP/1.1
Host: api.qubic.org
Accept: application/json
```
