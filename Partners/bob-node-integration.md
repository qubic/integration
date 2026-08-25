# Running your own Bob node for TX-based integration

## Table of contents

- [Overview](#overview)
- [1. Install Bob](#1-install-bob)
- [2. Configure Bob](#2-configure-bob)
    - [Peer discovery](#peer-discovery)
    - [Storage and retention](#storage-and-retention)
    - [Data retention model](#data-retention-model)
- [3. Maintain a Bob instance](#3-maintain-a-bob-instance)
    - [Epoch transition](#epoch-transition)
    - [Cold starts and re-syncing](#cold-starts-and-re-syncing)
    - [Version upgrades](#version-upgrades)
    - [Security considerations](#security-considerations)
    - [Troubleshooting](#troubleshooting)
- [4. Using Bob for TX-based integration](#4-using-bob-for-tx-based-integration)
    - [WebSocket subscriptions](#websocket-subscriptions)
    - [Broadcasting transactions](#broadcasting-transactions)
    - [Verifying transactions](#verifying-transactions)
    - [Useful REST endpoints](#useful-rest-endpoints)

## Overview

[Bob](https://github.com/qubic/core-bob) is a high-performance indexer for the Qubic network. It syncs tick data
from Qubic nodes, indexes it, and exposes it through an Ethereum-style **JSON-RPC 2.0 API** (HTTP and WebSocket)
and a **REST API**. It also broadcasts signed transactions to the network.

For partners and exchanges, Bob is an alternative to calling the public RPC (`https://rpc.qubic.org`): instead of
depending on Qubic's shared endpoints, you run a single Bob instance and point your integration at it. One
service covers the full exchange surface — deposit detection, transaction verification, asset data and
transaction broadcasting — replacing the equivalent assembled from the public RPC stack (Live API, Query API,
archiver and databases).

A self-hosted Bob provides:

- **Dedicated capacity** — no shared rate limits or contention on public endpoints.
- **Independent availability** — deposit crediting and withdrawals are unaffected by public RPC incidents.
- **Real-time push** — WebSocket subscriptions for deposit detection instead of polling.
- **Privacy** — deposit-address lookups and signed withdrawals do not transit a third-party service.
- **Locality** — Bob can be colocated with your own systems.

This guide is the operational companion to the [TX-based workflow](tx-based-workflow.md). That document describes
the integration logic (deposit scan, withdrawal, verification) against the public RPC; this one describes how to
run and operate the Bob infrastructure backing the same logic on your own node.

> [!IMPORTANT]
> The [Bob repository](https://github.com/qubic/core-bob) is the source of truth for installation, configuration
> and API details. This guide links to those documents rather than duplicating them, and adds only the
> operational guidance specific to exchange integration. Where the two differ, the repository takes precedence.

> [!TIP]
> For Bob-specific issues, configuration questions or incident support, use the dedicated Bob channel on the
> Qubic Discord: [#bob support](https://discord.com/channels/768887649540243497/1446161144044851210).

---

## 1. Install Bob

Installation instructions for Docker Hub, Docker from source, and native builds are in the
[Bob README](https://github.com/qubic/core-bob#quick-start). The documented minimum requirements are **4 cores
(AVX2), 12 GB RAM and 300 GB fast SSD/NVMe**.

> [!IMPORTANT]
> The 300 GB figure is a lower bound for running the service, not a capacity estimate. Actual disk consumption is
> determined by your retention setting (`kvrocks_ttl`). Observed epoch sizes reach **600–700 GB per epoch** at the
> high end, and epoch size varies considerably. Size the data volume as *(retained epochs) × (per-epoch size)*
> and see [Storage and retention](#storage-and-retention).

The Docker Hub image bundles Bob, Redis and Kvrocks in a single container. For production deployments, **pin a
version tag** (never `:latest`) and set a **restart policy** so Bob restarts automatically — this is required for
epoch transitions (see [Epoch transition](#epoch-transition)):

```bash
docker run -d --name qubic-bob --restart unless-stopped \
  -p 21842:21842 \
  -p 40420:40420 \
  -v qubic-bob-redis:/data/redis \
  -v qubic-bob-kvrocks:/data/kvrocks \
  -v qubic-bob-data:/data/bob \
  qubiccore/bob:<version>
```

Replace `<version>` with the current release tag from
[Docker Hub](https://hub.docker.com/r/qubiccore/bob) or the
[GitHub releases](https://github.com/qubic/core-bob/releases) page.

Port `21842` is the P2P server. Port `40420` serves the REST API and JSON-RPC (HTTP and WebSocket). Verify the
container is running with `docker logs -f qubic-bob`.

---

## 2. Configure Bob

Bob is configured through a single JSON file. When running the Docker image, the same fields can be set through
environment variables, so editing the JSON directly is usually unnecessary.

Configuration is documented across two reference pages in the Bob repository:
[CONFIG_FILE.MD](https://github.com/qubic/core-bob/blob/master/docs/CONFIG_FILE.MD) describes the JSON fields, and
[DOCKER_ENV.md](https://github.com/qubic/core-bob/blob/master/docs/DOCKER_ENV.md) describes the environment
variables and the bundled Redis/Kvrocks tuning. Some storage settings, including the retention TTL, appear only
in `DOCKER_ENV.md`.

### Peer discovery

Bob indexes the chain by syncing from other Qubic nodes. **In the default configuration it obtains peers
automatically** — no peer addresses need to be supplied.

**Automatic discovery (default).** When the `p2p-nodes` / `P2P_NODES` list is empty, Bob queries a peer-discovery
API at startup and requests a mix of lite (`BM:`) and Bob (`bob:`) peers. The connection pool continues to query
the same endpoints during operation, replacing the stalest connection one peer at a time, so the peer set
refreshes itself without intervention.

**Static peers.** If `p2p-nodes` is populated, startup discovery is skipped entirely and only the configured
endpoints are used. A `trusted-node` entry uses the format `BM:IP:PORT[:PASSCODE]`. Note that the connection pool
still calls discovery later when replacing a dead peer.

**Overriding the discovery endpoints.** The endpoint list is set by `peer_discovery_urls` in the config file or
`PEER_DISCOVERY_URLS` in Docker. Endpoints are tried in order and the **first non-empty response wins** — this is
failover, not aggregation.

```bash
PEER_DISCOVERY_URLS="https://api.qubic.global,https://api.qubic.li/public"
```

Format rules for each entry:

- **The scheme is required.** Without `://`, the entire string is treated as an origin with no path prefix and
  the HTTP client will fail to initialise.
- **Provide the base URL only.** Bob appends its own request path and query parameters.
- **A path prefix is preserved** — `https://api.qubic.li/public` resolves to `.../public/...`. Trailing slashes
  are trimmed.
- **A port may be included** in the origin, for example `http://10.0.0.5:8080`.
- **Surrounding whitespace is stripped** per element, so `"a, b"` is valid. Empty elements are discarded.

> [!NOTE]
> The compiled default is **two** endpoints — `https://api.qubic.global` followed by `https://api.qubic.li/public`.
> `DOCKER_ENV.md` currently documents a single default entry; the compiled default takes precedence.

Two further resources are fetched from config-driven URL lists at startup and are tried in order in the same way:
the current tick and epoch (`current_tick_endpoints`), and the spectrum/universe state files (`state_files_urls`).

### Storage and retention

Data flows **KeyDB (memory) → Kvrocks (disk)**. Each stage has a separate control. Confusing the two is the most
common cause of a disk incident, so the distinction matters:

| Field (env var)                           | What it controls                                                                                                                                                                                    | Guidance for exchanges                                                                                              |
|-------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------|
| `tick-storage-mode` (`TICK_STORAGE_MODE`) | `kvrocks` migrates old ticks to disk; `lastNTick` deletes them; `free` keeps everything in KeyDB                                                                                                    | Use `kvrocks` if history must survive restarts                                                                      |
| `tx-storage-mode` (`TX_STORAGE_MODE`)     | Same semantics for `transaction:*` blobs                                                                                                                                                            | Use `kvrocks` for durable transaction history                                                                       |
| `tx_tick_to_live` (`TX_TICK_TO_LIVE`)     | **Not a disk cap.** How many ticks of `transaction:*` blobs remain in **KeyDB before migrating to Kvrocks**                                                                                         | Tunes memory pressure and migration timing, not how long data survives on disk                                      |
| `kvrocks_ttl` (`KVROCKS_TTL`)             | **The disk-retention control.** TTL applied to every key Bob writes to Kvrocks (`transaction:*`, `itx:*`, `vtick:*`, `log:*`, indexed entries). Default `1209600` (**14 days**); `0` = never expire | Set to the smallest window covering your verification and reconciliation needs. `0` results in unbounded growth     |
| `last_n_tick_storage`                     | Ticks kept in memory in `lastNTick` mode                                                                                                                                                            | Bounds memory, not disk                                                                                             |
| `spam-qu-threshold`                       | Minimum QU amount to index a transfer                                                                                                                                                               | **Leave at `0`.** Raising it discards small transfers from the index, including genuine low-value customer deposits |

> [!WARNING]
> **Do not use `KVROCKS_ROCKSDB_TTL`.** It maps to `rocksdb.ttl`, which Kvrocks rejects as an invalid
> configuration key, so the setting silently has no effect. `DOCKER_ENV.md` currently describes it as a
> "RocksDB-level secondary TTL safety net"; that description is incorrect. The Bob team is aware and will correct
> the documentation. **`kvrocks_ttl` is the only functioning disk-retention control.**
>
> Verify with:
> ```bash
> docker exec qubic-bob grep -i "not a valid configuration key" /app/logs/kvrocks.log
> ```

**Sizing the retention window.** An epoch lasts seven days. A TTL of exactly 14 days does not reliably retain two
complete epochs, because expiry is relative to each key's write time rather than to epoch boundaries; a TTL of
**15 days** covers two full epochs with margin. Multiply the number of epochs you intend to retain by the
observed per-epoch size (up to ~700 GB at the high end) to size the volume.

**Preventing disk exhaustion:**

1. **Provision headroom.** Place the Kvrocks and Redis volumes on dedicated SSD/NVMe sized from your TTL, not
   from the 300 GB minimum.
2. **Set `kvrocks_ttl` deliberately.** This — not `tx_tick_to_live` — expires data from disk. Confirm it is not
   `0` unless unbounded retention is intended.
3. **Monitor and alert.** Treat free disk on the data volume as a first-class alerted metric; a full disk stalls
   or crashes the indexer. Monitor the `qubic-bob-kvrocks`, `qubic-bob-redis` and `qubic-bob-data` volumes.
4. **Verify overrides took effect.** At container start the entrypoint logs `=== Bob configuration ===`,
   `=== Redis configuration ===` and `=== Kvrocks configuration ===`. Confirm the TTL is applied rather than
   assuming it.
5. **Match storage mode to purpose.** If only recent data is required and a re-sync on restart is acceptable,
   `lastNTick` avoids unbounded disk growth.

### Data retention model

**Bob is a rolling window, not an archive, and is designed to be used as one.** Everything Bob writes to Kvrocks
expires after `kvrocks_ttl` — 14 days by default; non-durable storage modes retain less. The intended pattern is:

> **Read recent chain data from Bob, and store what matters in your own database.**

Bob provides a reliable real-time view of the most recent epochs. Your database is the system of record: as
deposits are detected and withdrawals confirmed, persist what your business requires (credited transfers,
transaction hashes, balances, audit trail) on your side. Data that ages out of Bob is no longer available from
your node. This is by design and is what keeps disk usage bounded.

In practice:

- **Persist as you process.** Write deposits and confirmations to your own store at the moment they are handled,
  rather than relying on re-reading them from Bob later.
- **Keep reconciliation lookback inside the retention window.** Any backfill or gap recovery performed against
  Bob (`POST /getQuTransfersForIdentity`, or a subscription with `startLogId`) must target data still within
  `kvrocks_ttl`. Raising the TTL extends the lookback at a proportional cost in disk.
- **Check what the node holds** with `GET /epochinfo/{epoch}`. It reports the epoch's `initialTick` and `endTick`
  and the node's `lastIndexedTick`; `found: false` indicates the epoch is no longer available locally.
- **For older history**, the public **Query API** (`/query/v1`) serves verified archived data — see the
  [official documentation](https://docs.qubic.org/apis/query#description/introduction) and the
  [TX-based workflow](tx-based-workflow.md#query-api). Treat it as a fallback for investigation, not as a
  substitute for your own records.

---

## 3. Maintain a Bob instance

### Epoch transition

Every Wednesday at approximately 12:00 UTC the Qubic network performs an **epoch transition**. Bob does not
simply stop and restart at this point; it executes a three-step transition. Interrupting that transition is the
most common cause of Bob data corruption. See also
[EPOCH_TRANSITION.MD](https://github.com/qubic/core-bob/blob/master/docs/EPOCH_TRANSITION.MD).

**Transition workflow:**

1. **Finalize.** Bob receives all `END_EPOCH` events, finalizes state, and stores the end-of-epoch logging events
   in a temporary location at `endTick + 1`.
2. **Wait window.** Bob continues serving the ended epoch for a configurable period — `wait_at_epoch_end`,
   default `1800` seconds (30 minutes) — so client integrations can complete their reading of that epoch's data.
   The window can be shortened or disabled in configuration.
3. **Migrate and restart.** Bob migrates its keys to a permanent location to avoid collision with the next epoch,
   then performs a clean exit so it can restart on the new epoch.

> [!WARNING]
> **Do not restart Bob during the wait window (step 2).** If Bob restarts before completing the migration in
> step 3, the temporary `endTick + 1` events of the old epoch collide with the first tick of the new epoch — both
> carry the same tick number — producing duplicated or misaligned data. A typical symptom is end-of-epoch
> transactions appearing as `zzz…unknowntransaction` instead of `SC_END_EPOCH_TX_<tick>`. Recovery generally
> requires wiping storage and re-syncing.

**Operator requirements:**

- **Keep an auto-restart supervisor** (`--restart unless-stopped`, a `systemd` unit, or an orchestrator). This is
  safe: a supervisor restarts Bob only after Bob's own clean exit in step 3, never during the window.
- **Do not manually restart, redeploy or upgrade Bob during the ~30-minute window after 12:00 UTC**, and ensure
  health and liveness probes tolerate it. A probe that kills the container because Bob is still serving the
  previous epoch causes exactly the corruption described above. Schedule upgrades well away from the transition.
- **Complete epoch-close reconciliation within the wait window.** The window exists so integrations can finish
  reading the ended epoch before Bob migrates it.
- **After the transition**, confirm Bob is on the new epoch and caught up via `GET /status`, and verify a past
  epoch's data with `GET /epochinfo/{epoch}`.

**Transactions in the virtual tick.** The virtual tick (`endTick + 1`) **cannot contain exchange or user
transactions**. It has no `TickData` or `TickVote` and never passes through consensus or quorum, and `TickData`
is what carries a tick's transaction set, so no user-signed transaction can be placed in it. The tick *number* is
reused: the next epoch's first real tick is also numbered `endTick + 1`, so a transaction targeting that number
is processed by the new epoch's real tick, never the virtual one. That number collision is why Bob migrates
virtual-tick data to a separate location, and why restarting mid-window corrupts data.

The virtual tick carries `END_EPOCH` events: logging events emitted while smart contracts run their end-of-epoch
procedures. These **can include `QU_TRANSFER` events** — for example smart-contract dividend or revenue
distributions and payouts, bracketed by the `STA_DDIV` / `END_DDIV` and `STA_EPOC` / `END_EPOC` markers. Bob
represents them as synthetic end-of-epoch transactions (`SC_END_EPOCH_TX_<tick>`), not user-signed transactions.

**Whether exchange deposit addresses are affected.** Generally not. End-of-epoch payouts target addresses holding
a position in the paying smart contract, such as a Qearn lock or dividend-bearing shares. A plain exchange
deposit address holds no such position, and Qubic guidance recommends against using client deposit accounts for
smart-contract activity (see the [deposit workflow](tx-based-workflow.md#deposit-workflow)). Under normal
operation these addresses are not end-of-epoch payout targets. It remains possible — for example a treasury or
hot wallet that does hold a smart-contract position, or a deposit address misused for smart-contract activity —
so handle it defensively:

- **If an end-of-epoch `QU_TRANSFER` credits one of your addresses, treat it like any other deposit** (gating on
  `executed` / settled status), even though no user-signed transaction backs it.
- These transfers belong to the **ending** epoch and are stored at `endTick + 1`, which is why Bob must not be
  restarted mid-window.

> [!NOTE]
> The set of smart contracts that move QU at end-of-epoch changes over time as contracts are added and updated.
> This guide does not enumerate them; the integration requirement is to handle any such transfer defensively
> rather than to track the list.

### Cold starts and re-syncing

When Bob starts from empty storage it must sync from its peers before its API reflects current network state.

**Cold-start duration depends on where the network is in the epoch**, because Bob syncs the current epoch from
its initial tick:

- Started shortly after an epoch transition, there are few ticks to process and sync can complete in minutes.
- Started late in an epoch, syncing the accumulated ticks can take **more than 24 hours**.

> [!IMPORTANT]
> **Do not schedule a cold start within roughly two days of an epoch transition** (Wednesday ~12:00 UTC). The
> sync is unlikely to complete before the epoch ends, so the instance will not capture the full epoch. Start cold
> instances early in an epoch instead.

Additional requirements:

- **Do not treat deposit or withdrawal state as authoritative until Bob has caught up to the current tick.**
- **Check sync state with `GET /status`.** The `SyncStatus` response reports `isSyncing`, `progress`, and
  `lastSeenNetworkTick` versus `currentIndexingTick`. Wait until Bob is no longer syncing before relying on its
  data.
- **Keep Kvrocks and Redis data on persistent volumes**, as in the install command, so a routine restart does not
  trigger a full re-sync.
- **For gap recovery after downtime**, WebSocket subscriptions support catch-up from a specific `logId`, allowing
  missed events to be replayed instead of rescanned — see [WebSocket subscriptions](#websocket-subscriptions).

### Version upgrades

Bob releases are published as Docker images and [GitHub releases](https://github.com/qubic/core-bob/releases).

1. **Read the [RELEASE_NOTES.md](https://github.com/qubic/core-bob/blob/master/RELEASE_NOTES.md)** for the target
   version. It lists user-visible RPC, REST, WebSocket and configuration changes, plus migration notes for
   breaking changes.
2. **Test the upgrade in a non-production environment first.** Confirm the new version syncs and that your
   integration behaves as expected, then promote the same pinned tag to production.
3. **Pull the new pinned image**, stop the old container, and start the new one against the **same data volumes**.
4. **Watch the logs** (`docker logs -f qubic-bob`) to confirm indexing resumes.
5. **Roll back** by re-pointing at the previously pinned tag if the release notes indicate an incompatibility.

Do not perform upgrades during the epoch transition window (see [Epoch transition](#epoch-transition)).

### Security considerations

Bob's exposure model depends on your intended deployment. Bob serves plain HTTP and WebSocket on port `40420`
with no built-in authentication, so decide deliberately which networks can reach it:

- **Restrict port `40420` to the networks that need it.** If the API is only consumed by your own systems, bind
  it to a private network rather than a public interface. Expose it publicly only if reachability by third
  parties is an explicit requirement.
- **Keep `enable-admin-endpoints: false`** (the default). Enable it only for a specific recovery operation and
  disable it afterwards — the admin endpoints can rewind and re-index the node.
- **Terminate TLS and apply authentication at a reverse proxy** if traffic to Bob crosses any untrusted hop.

### Troubleshooting

The two common failure modes are described below. Both are diagnosed primarily through `GET /status` and the
container logs. Data recovery uses Bob's **admin endpoints**, which exist only when `enable-admin-endpoints: true`
is set in configuration (off by default).

**Bob is not ticking (stalled).** Bob is running but its indexed tick does not advance.

- Check `GET /status` and compare `lastSeenNetworkTick` against `currentFetchingTick`,
  `currentFetchingLogTick` and `currentIndexingTick`. If the network tick advances but Bob's pointers do not, Bob
  is stalled — typically a peer or connectivity problem. If `lastSeenNetworkTick` is also static, Bob is not
  receiving data from the network at all.
- Check the logs (`docker logs -f qubic-bob`) for peer disconnects. Bob drops peers idle for approximately 30
  seconds, so an unhealthy peer set can starve it of data.
- Verify peer reachability. The repository ships `tools/bob_probe`, a CLI that probes a node's handshake,
  tick-info and log responses; use it to confirm configured or discovered peers respond.
- If peers are healthy, a restart can clear a stuck connection — but **never during the epoch wait window** (see
  [Epoch transition](#epoch-transition)).

**Bob is returning malformed or inconsistent data.** Responses contain garbage or misaligned data, such as
unexpected `zzz…unknowntransaction` entries or incorrect log ranges. This occurs when the computors or network
misbehave and Bob indexes a bad range of ticks, or after an interrupted epoch transition. The remedy is to rewind
and re-index from a point before the affected range:

1. **Locate the affected range** with `GET /_admin/checkIndexing?epoch=<epoch>` (or `?fromTick=&toTick=`) and
   `GET /_admin/checkTransactions?fromTick=&toTick=`, which report ticks with missing data, missing log ranges,
   or inconsistent transactions.
2. **Trigger a re-index** from a tick shortly before the corruption:
   ```bash
   curl -s -X POST http://<your-bob-host>:40420/_admin/reindex \
     -H "Content-Type: application/json" \
     -d '{"fromTick": <a_few_ticks_before_the_problem>}'
   ```
   This rewinds Bob's fetching, logging and indexing pointers to `fromTick` and reprocesses forward, rebuilding
   the affected data.
3. **Confirm recovery.** Watch `GET /status` until Bob has caught up, then re-check the previously affected ticks
   with `checkIndexing`.

**If a re-index does not resolve the problem** — for example corruption from an interrupted epoch transition, or
an instance left stuck after a non-seamless network restart — the fallback is to wipe Bob's storage and re-sync.
Bob normally handles epoch transitions and network restarts without manual intervention, so treat a wipe as a
last resort. **If retaining the existing data matters, ask in the
[#bob support Discord channel](https://discord.com/channels/768887649540243497/1446161144044851210) before
wiping**, and account for the cold-start timing constraints described in
[Cold starts and re-syncing](#cold-starts-and-re-syncing).

---

## 4. Using Bob for TX-based integration

Once Bob is running it replaces the public RPC for both sides of an exchange integration. The integration *logic*
is unchanged from the [TX-based workflow](tx-based-workflow.md) — deposit scan, withdrawal, verification, and the
[workflow guidelines](tx-based-workflow.md#qubic-workflow-guidelines) all continue to apply. What changes is the
endpoint: `http://<your-bob-host>:40420` instead of `https://rpc.qubic.org`.

**Porting from the public RPC to Bob.** Bob's API is not shaped like the Live and Query APIs, so calls must be
mapped:

| Public RPC (Live/Query)                                      | Bob equivalent                                                                                                                         |
|--------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------|
| `/live/v1/tick-info` (current tick, for target tick)         | `qubic_getTickNumber` (JSON-RPC)                                                                                                       |
| `/live/v1/broadcast-transaction`                             | `POST /broadcastTransaction` · `qubic_broadcastTransaction`                                                                            |
| `/query/v1/getLastProcessedTick` (how far data is available) | `GET /status` → `currentIndexingTick` · `qubic_syncing`                                                                                |
| `/query/v1/getTransactionByHash`                             | `GET /tx/{hash}` · `qubic_getTransactionByHash`                                                                                        |
| `/query/v1/getTransactionsForTick` (deposit scan)            | `GET /tick/{tick}` · `qubic_getTickByNumber` — or, preferably, a [WebSocket subscription](#websocket-subscriptions) instead of polling |
| `/query/v1/getProcessedTickIntervals`                        | `GET /epochinfo/{epoch}` · `qubic_getEpochInfo`                                                                                        |
| `/live/v1/balances/{identity}`                               | `GET /balance/{identity}` · `qubic_getBalance` (one identity per call)                                                                 |

> [!IMPORTANT]
> The Go examples in the TX-based workflow construct a `types.NewLiveServiceClient("https://rpc.qubic.org/live/v1")`
> and use it for both `GetTickInfo()` and `BroadcastTransaction()`. **That client does not communicate with Bob.**
> Continue using `go-node-connector` for wallet and transaction construction and signing — that logic is network
> format, not RPC, and is unchanged — but replace those two calls with Bob's endpoints per the table above.

### WebSocket subscriptions

Instead of polling tick by tick, subscribe over WebSocket and let Bob push matching events as they are indexed.
This is the recommended approach for deposit detection. Connect to `ws://<your-bob-host>:40420/ws/qubic` and use
the JSON-RPC methods `qubic_subscribe` and `qubic_unsubscribe`; a single connection can hold multiple
subscriptions. Full request and notification payloads are in
[QUBIC_JSON_RPC.md → WebSocket Subscriptions](https://github.com/qubic/core-bob/blob/master/docs/QUBIC_JSON_RPC.md#websocket-subscriptions).

**Subscription types.** Select the narrowest type that covers the requirement:

| Type         | Pushes                                                     | Catch-up           | Typical exchange use                                                                         |
|--------------|------------------------------------------------------------|--------------------|----------------------------------------------------------------------------------------------|
| `transfers`  | Settled QU transfers filtered by identity                  | Yes (`startLogId`) | **Deposit detection** — the lightest option                                                  |
| `logs`       | Log events matching an identity or log-type filter         | Yes (`startLogId`) | Asset transfers or specific smart-contract events for your addresses                         |
| `tickStream` | Every transaction and log per tick, with `executed` status | Yes (`startTick`)  | When execution status is required inline, or when building an equivalent of your own indexer |
| `newTicks`   | Tick metadata only (number, epoch, hash)                   | No                 | A lightweight "new tick arrived" signal; fetch details over REST                             |

**Connection lifecycle the client must handle:**

- **Subscribe and confirmation.** `qubic_subscribe` returns a subscription id (for example `qubic_sub_0`). Events
  then arrive as `qubic_subscription` messages carrying that id; `qubic_unsubscribe` cancels it.
- **Catch-up.** When `startLogId` (or `startTick`) is supplied, Bob replays missed events with `isCatchUp: true`,
  emits periodic `catchUpProgress` notifications, and sends `catchUpComplete` once the backlog is drained.
  Real-time events arriving during the replay are queued (up to **10,000**), so no events are dropped in between.
  After downtime, resubscribe from the saved cursor rather than rescanning over REST.
- **Heartbeat.** `tickStream` sends every 120th tick as a heartbeat, so a quiet stream still demonstrates
  liveness. Treat prolonged silence as a dead connection.
- **Reconnection is mandatory.** Bob exits at every [epoch transition](#epoch-transition), so the client must
  reconnect and re-subscribe automatically, resuming from its saved cursor. A client without a reconnect loop
  fails weekly.

#### Deposit detection with `transfers`

Subscribe to `transfers` filtered by your deposit identities:

```json
{
  "jsonrpc": "2.0",
  "method": "qubic_subscribe",
  "params": [
    "transfers",
    {
      "identity": [
        "YOUR_DEPOSIT_IDENTITY"
      ]
    }
  ],
  "id": 1
}
```

Each notification carries the epoch, tick, `logId`, `txHash` and a parsed body (`from`, `to`, `amount`) —
sufficient to credit the deposit.

> [!TIP]
> **To avoid missing deposits across restarts**, persist the last `logId` processed and, on reconnect,
> resubscribe with `startLogId` and `startEpoch`. Historical events replay with `isCatchUp: true`, followed by a
> `catchUpComplete` notification, after which real-time events resume.

The following example subscribes, credits, persists the cursor, and resumes from it on reconnect:

```javascript
const WebSocket = require('ws');

const BOB_WS = 'ws://<your-bob-host>:40420/ws/qubic';
const DEPOSIT_IDENTITIES = ['YOUR_DEPOSIT_IDENTITY_1', 'YOUR_DEPOSIT_IDENTITY_2'];

// Persist the cursor in your database, not in memory — it is the no-missed-deposit guarantee.
let lastLogId = loadLastProcessedLogId();   // null on first run
let lastEpoch = loadLastProcessedEpoch();

function connect() {
    const ws = new WebSocket(BOB_WS);

    ws.on('open', () => {
        const filter = {identity: DEPOSIT_IDENTITIES};
        // Resume from the last processed event: replays everything missed while disconnected.
        if (lastLogId !== null) {
            filter.startLogId = lastLogId + 1;
            filter.startEpoch = lastEpoch;
        }
        ws.send(JSON.stringify({
            jsonrpc: '2.0',
            method: 'qubic_subscribe',
            params: ['transfers', filter],
            id: 1,
        }));
    });

    ws.on('message', (data) => {
        const msg = JSON.parse(data);
        if (msg.id === 1 && msg.result) return;            // subscription confirmed
        if (msg.method !== 'qubic_subscription') return;

        const r = msg.params.result;
        if (r.catchUpProgress) return;                      // periodic progress during replay
        if (r.catchUpComplete) return;                      // backlog delivered; now live

        if (r.logTypename === 'QU_TRANSFER' && DEPOSIT_IDENTITIES.includes(r.body.to)) {
            // Must be idempotent: catch-up can replay events already credited.
            creditDepositOnce({
                dedupeKey: `${r.epoch}:${r.logId}`,
                identity: r.body.to,
                amount: r.body.amount,
                txHash: r.txHash,
                tick: r.tick,
            });
        }

        // Advance the cursor only after the deposit is durably credited.
        lastLogId = r.logId;
        lastEpoch = r.epoch;
        saveProgress(lastEpoch, lastLogId);
    });

    // Bob exits at every epoch transition, so always reconnect.
    ws.on('close', () => setTimeout(connect, 5000));
    ws.on('error', () => ws.close());
}

connect();
```

Two properties of this example are requirements for any implementation:

- **Idempotent crediting.** Catch-up replays events that may already have been processed, so key each credit on a
  stable identifier (`epoch` + `logId`) rather than crediting unconditionally.
- **Commit the cursor after the credit**, not before. A crash between the two otherwise loses a deposit silently.

> [!NOTE]
> `transfers` emits only QU transfers that actually settled, so no execution check is required — unlike
> `tickStream`, which also surfaces transactions that never executed.

#### Full tick stream

The `transfers` subscription pushes only settled QU-transfer log events. To receive every transaction per tick
together with its execution status, use the `tickStream` subscription. For each tick it returns the list of
transactions — each with an `executed` boolean and the log range it produced (`logIdFrom`, `logIdLength`) —
alongside the tick's log events. Narrow it with `txFilters` (`from`, `to`, `minAmount`, `inputType`) and catch up
from a past tick with `startTick`:

```json
{
  "jsonrpc": "2.0",
  "method": "qubic_subscribe",
  "params": [
    "tickStream",
    {
      "txFilters": [
        {
          "to": "YOUR_DEPOSIT_IDENTITY"
        }
      ]
    }
  ],
  "id": 1
}
```

> [!IMPORTANT]
> Unlike `transfers`, a `tickStream` transaction list also includes transactions that did **not** execute
> (`executed: false`). Always require `executed: true` before crediting a deposit. This is Bob's equivalent of the
> `moneyFlew` / `isExecuted` check in the [TX-based workflow](tx-based-workflow.md).

`tickStream` is more comprehensive — full transactions, logs and execution status in one stream — but heavier
than the specialised `transfers` filter. Full parameters and payloads are in
[QUBIC_JSON_RPC.md → WebSocket Subscriptions](https://github.com/qubic/core-bob/blob/master/docs/QUBIC_JSON_RPC.md#websocket-subscriptions).

### Broadcasting transactions

A self-hosted Bob broadcasts signed transactions, so the public Live API is not required for withdrawals. The
flow mirrors the [TX-based workflow](tx-based-workflow.md#creating-signing-sending-and-verifying-a-transaction)
(including the Qutil / Send-Many path) with two calls repointed at Bob.

**1. Get the current tick** to set a target tick in the future (the equivalent of the Live API's `tick-info`):

```bash
curl -s -X POST http://<your-bob-host>:40420/qubic \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"qubic_getTickNumber","params":[],"id":1}'
```

**2. Build and sign** the transaction with `go-node-connector` exactly as in the TX-based workflow. This step is
unchanged, since signing is network format rather than RPC. Set `targetTick` to the value from step 1 plus a
buffer; the TX-based workflow [recommends 15–20 ticks ahead](tx-based-workflow.md#respect-status-information) and
its examples use `+15`, so the transaction reaches the network before that tick passes. This is a network rule
and is unaffected by Bob.

**3. Serialize the signed transaction to hex.** This is the one substantive difference from the public Live API.
The workflow's `lsc.BroadcastTransaction(tx)` serializes and base64-encodes the transaction internally; Bob takes
the raw signed bytes as a **hex** string, so serialize it explicitly:

```go
tx, _ = signer.SignTx(tx) // signed transaction (from step 2)
packet, _ := tx.MarshallBinary() // full signed bytes, signature included
data := hex.EncodeToString(packet) // Bob requires hex
```

> [!WARNING]
> Bob's `data` field is **hex**, not base64. `go-node-connector`'s `EncodeToBase64()`, used by the public Live
> API, produces base64 and must not be sent to Bob. Use `MarshallBinary()` with `hex.EncodeToString`.

**4. Broadcast** the hex payload (optional `0x` prefix, even length) to your node:

```bash
curl -s -X POST http://<your-bob-host>:40420/broadcastTransaction \
  -H "Content-Type: application/json" \
  -d '{"data":"<signed-transaction-hex>"}'
```

The response is a `BroadcastResult` (`success`, `txHash`, `error`). The equivalent JSON-RPC method is
`qubic_broadcastTransaction`. See the
[REST API reference](https://github.com/qubic/core-bob/blob/master/docs/REST_API.md) for the full endpoint
contract.

> [!NOTE]
> Bob's documentation currently covers `broadcastTransaction` with an already-signed hex payload; a complete
> build → sign → broadcast example will be added to the Bob repository. Until then, the sequence above is the
> reference for the hex serialization step.

> [!NOTE]
> The one-concurrent-transaction-per-source-address and target-tick rules from the
> [workflow guidelines](tx-based-workflow.md#qubic-workflow-guidelines) still apply; they are network rules, not
> RPC-specific. The target tick must be derived from **step 1 (Bob's live view)**, not from `GET /status`'s
> `currentIndexingTick`, which reports how far Bob has indexed and lags the network.

### Verifying transactions

After broadcasting, verify execution and retrieve the resulting log events via Bob's REST API:
`GET /tx/{tx_hash}` returns `isExecuted`, `logIdFrom` and `logIdTo`, then `GET /log/{epoch}/{from}/{to}` returns
the events. This flow is documented in
[DEAL_WITH_TX.MD](https://github.com/qubic/core-bob/blob/master/docs/DEAL_WITH_TX.MD).

### Useful REST endpoints

WebSocket streams are used to react to new activity; the REST API is used for direct queries — health checks,
balance lookups, one-off verification, and retrieving historical ranges for reconciliation without replaying a
stream. The full contract is in
[REST_API.md](https://github.com/qubic/core-bob/blob/master/docs/REST_API.md). The endpoints most relevant to an
exchange:

| Endpoint                                                      | Purpose                                                                                                                   |
|---------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------|
| `GET /status`                                                 | Node sync and health (`isSyncing`, `progress`, indexing versus network tick). Use for monitoring and to confirm catch-up. |
| `GET /epochinfo/{epoch}`                                      | The tick range the node holds for an epoch, and whether it holds that epoch at all (`found`).                             |
| `GET /balance/{identity}`                                     | Current balance of an identity; used for reconciliation and withdrawal pre-checks.                                        |
| `GET /tx/{tx_hash}`                                           | Verify a specific transaction (`isExecuted`, log range).                                                                  |
| `GET /log/{epoch}/{from_id}/{to_id}`                          | Fetch the log events a transaction produced.                                                                              |
| `POST /getQuTransfersForIdentity`                             | QU transfers for a deposit identity within a tick range; historical backfill and gap reconciliation.                      |
| `POST /getAssetTransfersForIdentity`                          | Asset transfers for an identity (by issuer and asset name) within a tick range.                                           |
| `GET /asset/{identity}/{issuer}/{asset_name}/{manageSCIndex}` | Asset balance for an identity.                                                                                            |
| `POST /broadcastTransaction`                                  | Send a signed transaction (see [Broadcasting transactions](#broadcasting-transactions)).                                  |
| `POST /querySmartContract`                                    | Query smart-contract state.                                                                                               |

> [!TIP]
> A common pattern is to subscribe over WebSocket for real-time deposit detection and use
> `POST /getQuTransfersForIdentity` over a tick range to reconcile or backfill after downtime, bounded by the
> [retention window](#data-retention-model).

---

The canonical documentation for Bob lives in the [Bob repository](https://github.com/qubic/core-bob), linked
throughout this guide. See also the [Partner integration overview](README.md).
