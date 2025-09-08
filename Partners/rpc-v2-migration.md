# Migrating to RPC V2
***
## Introduction
The aim for RPC V2 is to improve the performance and the API interfaces.   
In order to facilitate this, the underlying infrastructure underwent massive changes, which you can read more about [here](https://qubic.org/blog-detail/rpc-2-0-qubic-integration-layer-functionality-upgrade).  
> **The old infrastructure will eventually be sunset. We recommend that you migrate your services to use the new infrastructure.**  
> **Do not hesitate to reach out to us on Discord if you encounter issues using the new version. An invitation link can be found on the [Qubic website](https://qubic.org/).**
***

## Usage

### Domain change

Currently, the V2 infrastructure is available at `https://api.qubic.org/`, while the old version is still available at `https://rpc.qubic.org/`.  

### Endpoint changes

1. The first big change is that the naming convention of the endpoints has changed to resemble RPC calls.  
**Example:** `https://rpc.qubic.org/v1/ticks/00000000/tick-data` => `https://api.qubic.org/getTickData`

2. The second and more important change is that endpoint input is no longer passed via path and query parameters.  
Instead it is passed as JSON in the request body. This change provides improved flexibility when using filters, tick ranges and pagination options.
> A secondary effect of this change is that endpoints which require input, are now made via the **POST** method.  

**Example: Getting an identity's transactions**

RPC V1 provided multiple endpoints related to identity transactions, thus some filtering came up to what endpoint was used to query.  
RPC V2 allows for a variety of options when querying transactions related information, in a more expressive way.

Querying in V1: 
```bash
curl https://rpc.qubic.org/v2/identities/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFXIB/transfers?startTick=32302000&endTick=32302119 | jq  
```

Querying in V2: 
```bash
curl \
 -X POST \
 -H 'Content-Type: application/json'\
 -d '{
   "identity": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFXIB",
   "ranges": {
     "tickNumber": {
      "gte": "32302000",
      "lte": "32302119"
     },
     "amount": {
      "lt": "10000"
     }
   },
   "filters": {
    "source": "IOGWKHJWJTBOZAACKOQKRBZTLWWAMGFXNVHGMNCGNGOJPILZUIDEIZADIFVN",
    "inputType": "8"
   },
   "pagination": {
    "offset": 0,
    "size": 2
   }
 }' \
 'https://api.qubic.org/getTransactionsForIdentity' \
 | jq
```





