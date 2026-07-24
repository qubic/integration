# Running your own Bob node for TX-based integration

> **Status: DRAFT** — being prepared for partners/exchanges (see
> [issue #113](https://github.com/qubic/integration/issues/113)). Content and links are still under review, and
> a few items are still open (see [Open questions](#open-questions-to-resolve-before-publishing)).

## Table of contents

- [Overview](#overview)
- [1. Install Bob](#1-install-bob)
- [2. Configure Bob](#2-configure-bob)
  - [Data retention](#data-retention)
- [3. Maintain a Bob instance](#3-maintain-a-bob-instance)
  - [Epoch transition](#epoch-transition)
  - [Cold starts and re-syncing](#cold-starts-and-re-syncing)
  - [Version upgrades](#version-upgrades)
  - [Troubleshooting](#troubleshooting)
- [4. Using Bob for TX-based integration](#4-using-bob-for-tx-based-integration)
  - [WebSocket subscriptions](#websocket-subscriptions)
  - [Broadcasting transactions](#broadcasting-transactions)
  - [Verifying transactions](#verifying-transactions)
  - [Useful REST endpoints](#useful-rest-endpoints)
- [Open questions to resolve before publishing](#open-questions-to-resolve-before-publishing)

## Overview

[Bob](https://github.com/qubic/core-bob) is a high-performance indexer for the Qubic network. It syncs tick data
directly from Qubic nodes, indexes it, and exposes it through an Ethereum-style **JSON-RPC 2.0 API** (HTTP &
WebSocket) and a **REST API**. It can also **broadcast signed transactions** to the network.

For partners and exchanges, Bob is an **alternative to calling the public RPC** (`https://rpc.qubic.org`):
instead of depending on Qubic's shared endpoints, you run one Bob instance and point your integration at it. It
covers what an exchange needs — deposit detection, transaction verification, asset data and transaction
broadcasting — in a single service. Building the equivalent from the public RPC stack (Live API + Query API +
archiver + databases) is the effort we want to spare you; Bob is one container instead.

Compared with calling the public RPC, a self-hosted Bob gives you:

- **Dedicated capacity** — no shared rate limits or contention on the public endpoint.
- **Availability under your own control** — deposit crediting and withdrawals don't stop if the public RPC has an incident.
- **Real-time push** — WebSocket subscriptions for deposit detection instead of polling.
- **Privacy** — deposit-address lookups and signed withdrawals don't transit a third-party service.
- **Locality** — Bob can be colocated with your own systems.

> [!NOTE]
> Bob is not self-contained: it indexes the chain by syncing from **Qubic node(s) that you configure**
> (`trusted-node` / `p2p-nodes`). Where an exchange gets those node addresses — an official list, dedicated
> peers, or a node it runs itself — is the
> [key open question](#open-questions-to-resolve-before-publishing) in this draft.

This guide is the **operational companion** to the [TX-based workflow](tx-based-workflow.md): that document
describes the integration logic (deposit scan, withdrawal, verification) against the public RPC; this one
describes how to run the Bob infrastructure that backs the same logic on your own node.

> [!IMPORTANT]
> The [Bob repository](https://github.com/qubic/core-bob) is the **source of truth** for installation,
> configuration and API details. This guide intentionally links to those documents rather than copying them, and
> only adds the operational guidance specific to exchange integration. Where the two differ, trust the repository.

> [!TIP]
> **Need help?** For Bob-specific issues, doubts, or incident support, use the Qubic Discord's dedicated Bob
> channel: [#bob support](https://discord.com/channels/768887649540243497/1446161144044851210).

> [!NOTE]
> **Open questions to resolve before publishing** are collected [at the end](#open-questions-to-resolve-before-publishing).
> The most important one for exchanges: **who provides the trusted peer nodes to sync from?**

---

## 1. Install Bob

Full instructions — Docker Hub, Docker from source, and native build — are in the
[Bob README](https://github.com/qubic/core-bob#quick-start). Minimum requirements are **4 cores (AVX2), 12 GB RAM
and 300 GB fast SSD/NVMe**.

The fastest path is the Docker Hub image, which bundles Bob, Redis and Kvrocks in one container. For production,
**pin a version tag** (not `:latest`) and set a **restart policy** so Bob comes back automatically — this is
required for epoch transitions (see [below](#epoch-transition)):

```bash
docker run -d --name qubic-bob --restart unless-stopped \
  -p 21842:21842 \
  -p 40420:40420 \
  -v qubic-bob-redis:/data/redis \
  -v qubic-bob-kvrocks:/data/kvrocks \
  -v qubic-bob-data:/data/bob \
  qubiccore/bob:1.5.x
```

Port `21842` is the P2P server; port `40420` serves the REST API and JSON-RPC (HTTP + WebSocket). Confirm it is
running with `docker logs -f qubic-bob`.

---

## 2. Configure Bob

Bob is configured through a single JSON file. If you run the Docker image, environment variables can set those
same fields, so you rarely edit the JSON directly.

The Bob repository documents the available settings across two reference pages, and it is worth knowing both:
[CONFIG_FILE.MD](https://github.com/qubic/core-bob/blob/master/docs/CONFIG_FILE.MD) describes the JSON fields,
while [DOCKER_ENV.md](https://github.com/qubic/core-bob/blob/master/docs/DOCKER_ENV.md) describes the
environment variables and the bundled Redis/Kvrocks tuning. Some storage settings — including the retention TTL
below — are only described in `DOCKER_ENV.md`.

This section covers the settings that matter most for exchange operators, especially the problem the public infra
is meant to spare you: **running out of disk**.

Data flows **KeyDB (memory) → Kvrocks (disk)**, and each stage has its own knob. Getting these two confused is
the usual cause of a disk incident, so be precise about which does what:

| Field (env var) | What it actually controls | Guidance for exchanges |
| --------------- | ------------------------- | ---------------------- |
| `tick-storage-mode` (`TICK_STORAGE_MODE`) | `kvrocks` migrates old ticks to disk; `lastNTick` deletes them; `free` keeps everything in KeyDB | Use `kvrocks` if you need history to survive restarts |
| `tx-storage-mode` (`TX_STORAGE_MODE`) | Same semantics for `transaction:*` blobs | Use `kvrocks` for durable transaction history |
| `tx_tick_to_live` (`TX_TICK_TO_LIVE`) | **Not a disk cap.** How many ticks of `transaction:*` blobs stay in **KeyDB before migrating to Kvrocks** | Tunes **memory** pressure and migration timing — *not* how long data survives on disk |
| `kvrocks_ttl` (`KVROCKS_TTL`) | **The actual disk-retention knob.** TTL applied to every key Bob writes to Kvrocks (`transaction:*`, `itx:*`, `vtick:*`, `log:*`, indexed entries). Default `1209600` (**14 days ≈ 2 epochs**); `0` = **never expire** | Set to the smallest window covering your verification/reconciliation needs. `0` means unbounded growth |
| `last_n_tick_storage` | Ticks kept in memory in `lastNTick` mode | Bounds memory, not disk |
| `spam-qu-threshold` | Minimum QU amount to index a transfer | Raise to skip dust/spam and shrink the index |

**To avoid running out of disk:**

1. **Provision headroom** — put the Kvrocks/Redis volumes on a dedicated SSD/NVMe well above the 300 GB floor.
2. **Set `kvrocks_ttl` deliberately** — this, not `tx_tick_to_live`, is what expires data from disk. Confirm it is
   **not `0`** unless you genuinely intend to keep everything forever.
3. **Filter dust** — set a sensible `spam-qu-threshold`. <!-- TODO: recommended default? (see open questions) -->
4. **Monitor and alert** — treat free disk on the data volume as a first-class alerted metric; a full disk will
   stall or crash the indexer. Watch the `qubic-bob-kvrocks` / `qubic-bob-redis` / `qubic-bob-data` volumes.
5. **Verify your overrides took effect** — at container start the entrypoint logs `=== Bob configuration ===`,
   `=== Redis configuration ===` and `=== Kvrocks configuration ===`. Check your TTL is actually applied rather
   than assuming it.
6. **Match mode to purpose** — if you only need recent data and can tolerate a re-sync on restart, `lastNTick`
   avoids unbounded disk growth.

> [!NOTE]
> Bob must sync from at least one **trusted peer node** (`trusted-node`, format `BM:IP:PORT[:PASSCODE]`) or
> `p2p-nodes`. See the [open questions](#open-questions-to-resolve-before-publishing) — we need to confirm how
> exchanges obtain these.

### Data retention

**Bob is a rolling window, not an archive — and it is meant to be used that way.** Everything Bob writes to
Kvrocks expires after `kvrocks_ttl`, **14 days (~2 epochs) by default** (non-durable storage modes keep even
less). The intended pattern is:

> **Read recent chain data from Bob → store what matters in your own database.**

Bob's job is to give you a reliable, real-time view of the last couple of epochs. Your database is your **system
of record**: as you detect deposits and confirm withdrawals, persist what your business needs (credited
transfers, tx hashes, balances, audit trail) on your side. Data that ages out of Bob is gone from your node — by
design, not as a limitation, and it is what keeps disk usage bounded.

What this means in practice:

- **Persist as you process.** Write deposits and confirmations to your own store at the moment you handle them,
  rather than planning to re-read them from Bob later.
- **Keep your reconciliation lookback inside the window.** Any backfill or gap-recovery you do against Bob (e.g.
  `POST /getQuTransfersForIdentity`, or a subscription with `startLogId`) must target data still inside
  `kvrocks_ttl`. Raise the TTL deliberately if you need a longer lookback — and pay for it in disk.
- **Check what your node actually holds** with `GET /epochinfo/{epoch}` — it reports the epoch's `initialTick` /
  `endTick` and the node's `lastIndexedTick`, and `found: false` means that epoch is no longer available locally.
- **For older history you did not keep**, the public **Query API** (`/query/v1`) serves verified archived data —
  see its [official documentation](https://docs.qubic.org/apis/query#description/introduction) and the
  [TX-based workflow](tx-based-workflow.md#query-api). Treat it as a fallback for investigation, not as a
  substitute for your own records.

---

## 3. Maintain a Bob instance

### Epoch transition

Every Wednesday at ~12:00 UTC the Qubic network undergoes an **epoch transition**. Bob does **not** simply stop
and restart at this moment — it runs a three-step transition, and interrupting it is the single most common way
operators corrupt their Bob data. See also
[EPOCH_TRANSITION.MD](https://github.com/qubic/core-bob/blob/master/docs/EPOCH_TRANSITION.MD).

**Bob's transition workflow:**

1. **Finalize.** Bob receives all `END_EPOCH` events, finalizes the state, and stores the end-of-epoch logging
   events in a temporary space at `endTick + 1`.
2. **Serve a wait window.** Bob keeps serving the just-ended epoch for a configurable period —
   `wait_at_epoch_end`, **default `1800` seconds (30 min)** — so third-party apps (i.e. **your** integration) can
   finalize their reading of that epoch's data. Can be shortened or disabled in config.
3. **Migrate & restart.** Bob migrates its keys to a permanent location (to avoid duplication with the next
   epoch), then performs its clean exit so it can come back up on the new epoch.

> [!WARNING]
> **Do not restart Bob during the wait window (step 2).** If Bob is restarted before it completes the migration
> in step 3, the temporary `endTick + 1` events of the old epoch **collide with the first tick of the new epoch**
> (they share the same tick number), producing duplicated / misaligned data and symptoms such as end-of-epoch
> transactions showing up as `zzz…unknowntransaction` instead of `SC_END_EPOCH_TX_<tick>`. Recovering usually
> means wiping and re-syncing.

**What the operator must do:**

- **Keep the auto-restart supervisor** (the `--restart unless-stopped` policy, a `systemd` service, or an
  orchestrator). This is safe: a supervisor only restarts Bob *after Bob's own clean exit in step 3*, never
  during the window.
- **Never manually restart, redeploy, or upgrade Bob during the ~30-minute post-12:00-UTC window**, and make
  sure any **health/liveness probes tolerate it** — a probe that kills the container because Bob is "still on the
  old epoch" will trigger exactly the corruption above. Schedule upgrades well away from the transition.
- **Do your own epoch-close reconciliation within the wait window** — that window exists precisely so your
  integration can finish reading the ended epoch before Bob migrates it.
- **After the transition**, confirm Bob is on the new epoch and caught up via `GET /status`, and verify a past
  epoch's data with `GET /epochinfo/{epoch}` (returns the correct log range for that epoch).

**What transactions to expect in the window.** The virtual tick (`endTick + 1`) **cannot contain exchange or
user transactions.** By construction it has no `TickData` or `TickVote` and never passes through consensus/quorum
— and `TickData` is exactly what carries a tick's transaction set, so no user-signed transaction can be placed in
it. (The tick *number* is reused: the next epoch's first *real* tick is also numbered `endTick + 1`, so a
transaction that targets that number is processed by the new epoch's real tick, never the virtual one. That
number collision is why Bob migrates the virtual-tick data to a separate place, and why restarting mid-window
corrupts data.)

What the virtual tick *does* carry are the `END_EPOCH` events: logging events emitted while smart contracts run
their end-of-epoch procedures. These **can include `QU_TRANSFER` events** — e.g. smart-contract dividend /
revenue distributions and payouts (bracketed by the `STA_DDIV` / `END_DDIV` and `STA_EPOC` / `END_EPOC` markers).
Bob represents them as synthetic end-of-epoch transactions (`SC_END_EPOCH_TX_<tick>`), not user-signed ones.

**Are exchange deposit addresses normally targeted by these?** Generally **no.** End-of-epoch payouts go to
addresses that hold a position in the paying smart contract (e.g. a Qearn lock, or shares that earn a dividend).
A plain exchange deposit address holds no such position — and Qubic guidance already recommends **not** using
client deposit accounts for smart-contract activity (see the
[deposit workflow](tx-based-workflow.md#deposit-workflow)) — so under normal operation these addresses are not
end-of-epoch payout targets. It *can* still happen (e.g. a treasury/hot wallet that does hold an SC position, or
a deposit address misused for SC activity), so handle it defensively:

- **Don't assume, but don't expect.** If an end-of-epoch `QU_TRANSFER` does credit one of your addresses, treat
  it like any other deposit (gating on `executed` / settled), even though there is no user-signed transaction
  behind it.
- These transfers belong to the **ending** epoch and are stored at `endTick + 1` — which is exactly why Bob must
  not be restarted mid-window (otherwise they collide with the next epoch's first tick, per the warning above).

> [!NOTE]
> Which specific smart contracts move QU at end-of-epoch changes over time (contracts get added and updated), so
> this guide deliberately does not enumerate them — the integration point is simply to handle any such transfer
> defensively, not to track the list.

### Cold starts and re-syncing

When Bob starts from empty storage it must sync from its peers before its API reflects the current network state.

- Expect a catch-up period after a cold start; the node is behind until it finishes. **Do not treat deposit or
  withdrawal state as authoritative until Bob has caught up to the current tick.**
- Check sync state with **`GET /status`** — the `SyncStatus` response reports `isSyncing`, `progress`, and
  `lastSeenNetworkTick` vs `currentIndexingTick`. Wait until Bob is no longer syncing before relying on its data.
- Keep the KVRocks/Redis data on **persistent volumes** (as in the install command) so a routine restart does
  *not* trigger a full re-sync.
- For gap recovery after downtime, WebSocket subscriptions support **catch-up from a specific `logId`** so you
  can replay missed events instead of rescanning — see [WebSocket subscriptions](#websocket-subscriptions).

<!-- TODO: expected cold-start / full-resync duration (see open questions) -->

### Version upgrades

Bob releases are published as Docker images and [GitHub releases](https://github.com/qubic/core-bob/releases).

1. **Read the [RELEASE_NOTES.md](https://github.com/qubic/core-bob/blob/master/RELEASE_NOTES.md)** for the target
   version — it lists user-visible RPC/REST/WS/config changes and migration notes for breaking changes.
2. **Test the upgrade in staging/development first.** Run the new version against a non-production Bob instance,
   confirm it syncs and that your integration still behaves as expected, then promote the **same pinned tag** to
   production. Never upgrade production directly.
3. **Pull the new image** (pinned, e.g. `qubiccore/bob:1.5.x`), stop the old container, start the new one against
   the **same data volumes**.
4. **Watch the logs** (`docker logs -f qubic-bob`) to confirm it resumes indexing.
5. **Roll back** by re-pointing at the previously pinned tag if the release notes flag an incompatibility.

### Troubleshooting

Two failure modes worth having a runbook for. Both are diagnosed primarily through `GET /status` and the logs.
Data recovery uses Bob's **admin endpoints**, which exist only when `enable-admin-endpoints: true` is set in
config (**off by default** — enable it deliberately for recovery and never expose it publicly).

**Bob is not ticking (stalled).** Bob is running but its indexed tick stops advancing.

- Check `GET /status`: compare `lastSeenNetworkTick` against `currentFetchingTick` / `currentFetchingLogTick` /
  `currentIndexingTick`. If the network tick advances but Bob's pointers don't, Bob is stalled — usually a
  peer/connectivity problem. If even `lastSeenNetworkTick` is stuck, Bob isn't hearing from the network at all.
- Check the logs (`docker logs -f qubic-bob`) for peer disconnects. Bob drops peers idle for ~30s, so an
  unhealthy `trusted-node` / `p2p-nodes` set can starve it of data.
- Verify your peers are reachable and responsive. The repo ships `tools/bob_probe`, a CLI to probe a node's
  handshake / tick-info / log responses — use it to confirm your configured peers actually answer.
- If peers are healthy, a restart can clear a stuck connection — but **never during the epoch wait window**
  (see [Epoch transition](#epoch-transition)).

**Bob is returning malformed or inconsistent data.** Responses contain garbage or misaligned data (e.g.
unexpected `zzz…unknowntransaction` entries, wrong log ranges). This happens when the computors/network misbehave
and Bob indexes a bad stretch of ticks, or after an interrupted epoch transition. Per the Bob maintainers, the
fix is to **rewind and re-index from a few ticks before the problem**:

1. **Locate the bad range** with `GET /_admin/checkIndexing?epoch=<epoch>` (or `?fromTick=&toTick=`) and
   `GET /_admin/checkTransactions?fromTick=&toTick=`, which report ticks with missing data, missing log ranges,
   or inconsistent transactions.
2. **Trigger a re-index** from a tick shortly *before* the corruption:
   ```bash
   curl -s -X POST http://<your-bob-host>:40420/_admin/reindex \
     -H "Content-Type: application/json" \
     -d '{"fromTick": <a_few_ticks_before_the_problem>}'
   ```
   This rewinds Bob's fetching / logging / indexing pointers to `fromTick` and reprocesses forward, rebuilding
   the affected data.
3. **Confirm recovery** — watch `GET /status` until Bob has caught back up, then re-check the previously bad
   ticks with `checkIndexing`.

> [!NOTE]
> If a re-index does not clear it — for example corruption from an interrupted
> [epoch transition](#epoch-transition) — the fallback is to wipe Bob's storage and re-sync. When in doubt, ask
> in the [#bob support Discord channel](https://discord.com/channels/768887649540243497/1446161144044851210).

---

## 4. Using Bob for TX-based integration

Once Bob is running, it replaces the public RPC for both sides of an exchange integration. The integration
*logic* is unchanged from the [TX-based workflow](tx-based-workflow.md) — deposit scan, withdrawal, verification,
and the [workflow guidelines](tx-based-workflow.md#qubic-workflow-guidelines) all still apply. What changes is
*where* you call: `http://<your-bob-host>:40420` instead of `https://rpc.qubic.org`.

**Porting from the public RPC to Bob.** Bob's API is not shaped like the Live/Query APIs, so calls must be
mapped. If you already follow the [TX-based workflow](tx-based-workflow.md), this table is the translation:

| Public RPC (Live/Query) | Bob equivalent |
| ----------------------- | -------------- |
| `/live/v1/tick-info` (current tick, for target tick) | `qubic_getTickNumber` (JSON-RPC) |
| `/live/v1/broadcast-transaction` | `POST /broadcastTransaction` · `qubic_broadcastTransaction` |
| `/query/v1/getLastProcessedTick` (how far data is available) | `GET /status` → `currentIndexingTick` · `qubic_syncing` |
| `/query/v1/getTransactionByHash` | `GET /tx/{hash}` · `qubic_getTransactionByHash` |
| `/query/v1/getTransactionsForTick` (deposit scan) | `GET /tick/{tick}` · `qubic_getTickByNumber` — or, preferably, a [WebSocket subscription](#websocket-subscriptions) instead of polling |
| `/query/v1/getProcessedTickIntervals` | `GET /epochinfo/{epoch}` · `qubic_getEpochInfo` |
| `/live/v1/balances/{identity}` | `GET /balance/{identity}` · `qubic_getBalance` (one identity per call) |

> [!NOTE]
> **Bob does not replace the Stats API.** The `/v1/latest-stats` endpoint serves market aggregates — price,
> market cap, circulating supply, active addresses, burned QU. These are informational, not part of the
> deposit/withdraw flow, and Bob does not produce them; keep using the public Stats API. One field, **`epochTickQuality`**,
> is chain-derived rather than market data and *could* be served by Bob but is not exposed today — so it too
> currently requires the Stats API (see [open question #8](#open-questions-to-resolve-before-publishing)). See the
> [Partner integration overview](README.md) for where the Stats API fits.

> [!IMPORTANT]
> The Go examples in the TX-based workflow build a `types.NewLiveServiceClient("https://rpc.qubic.org/live/v1")`
> and use it for both `GetTickInfo()` and `BroadcastTransaction()`. **That client does not talk to Bob.** Keep
> using `go-node-connector` for wallet/transaction construction and signing (that logic is unchanged — it is
> network format, not RPC), but replace those two calls with Bob's endpoints per the table above.

### WebSocket subscriptions

Instead of polling tick by tick, subscribe over WebSocket and let Bob **push** matching events as they are
indexed. This is the recommended approach for deposit detection. Connect to
`ws://<your-bob-host>:40420/ws/qubic` and use the JSON-RPC methods `qubic_subscribe` / `qubic_unsubscribe`; a
single connection can hold several subscriptions at once. Full request/notification payloads are in
[QUBIC_JSON_RPC.md → WebSocket Subscriptions](https://github.com/qubic/core-bob/blob/master/docs/QUBIC_JSON_RPC.md#websocket-subscriptions).

**Subscription types** — pick the narrowest one that covers your need:

| Type | Pushes | Catch-up | Typical exchange use |
| ---- | ------ | -------- | -------------------- |
| `transfers` | Settled QU transfers filtered by identity | Yes (`startLogId`) | **Deposit detection** — the lightest option |
| `logs` | Log events matching an identity / log-type filter | Yes (`startLogId`) | Asset transfers or specific SC events for your addresses |
| `tickStream` | Every transaction + log per tick, with `executed` status | Yes (`startTick`) | When you need execution status inline or are effectively building your own indexer |
| `newTicks` | Tick metadata only (number, epoch, hash) | No | A lightweight "new tick arrived" signal; fetch details over REST |

**Connection lifecycle you must handle:**

- **Subscribe → confirmation.** `qubic_subscribe` returns a subscription id (e.g. `qubic_sub_0`). Events then
  arrive as `qubic_subscription` messages carrying that id; `qubic_unsubscribe` cancels it.
- **Catch-up (your gap-recovery mechanism).** When you pass `startLogId` (or `startTick`), Bob first **replays**
  missed events with `isCatchUp: true`, emits periodic `catchUpProgress`, and sends `catchUpComplete` once the
  backlog is drained. Real-time events that arrive during the replay are queued (up to **10,000**) so nothing is
  dropped in between. After downtime, resubscribe from your saved cursor rather than rescanning over REST.
- **Heartbeat.** A quiet stream still proves it is alive — `tickStream` sends every 120th tick as a heartbeat.
  Treat prolonged silence as a dead connection.
- **Reconnect is mandatory.** Bob exits at every [epoch transition](#epoch-transition), so your client must
  reconnect and re-subscribe automatically, resuming from your saved cursor (see the example below). A client
  without a reconnect loop dies every week.

#### Deposit detection with `transfers`

Subscribe to `transfers` filtered by your deposit identities:

```json
{
  "jsonrpc": "2.0",
  "method": "qubic_subscribe",
  "params": ["transfers", { "identity": ["YOUR_DEPOSIT_IDENTITY"] }],
  "id": 1
}
```

Each notification carries the epoch, tick, `logId`, `txHash` and a parsed body (`from`, `to`, `amount`) — enough
to credit the deposit.

> [!TIP]
> **Never miss a deposit across restarts:** persist the last `logId` you processed, and on reconnect resubscribe
> with `startLogId` (and `startEpoch`). Historical events replay with `isCatchUp: true`, followed by a
> `catchUpComplete` notification, then real-time events resume.

Putting those pieces together — subscribe, credit, persist the cursor, and resume from it on reconnect:

```javascript
const WebSocket = require('ws');

const BOB_WS = 'ws://<your-bob-host>:40420/ws/qubic';
const DEPOSIT_IDENTITIES = ['YOUR_DEPOSIT_IDENTITY_1', 'YOUR_DEPOSIT_IDENTITY_2'];

// Persist the cursor in your database, not in memory — it is your no-missed-deposit guarantee.
let lastLogId = loadLastProcessedLogId();   // null on first run
let lastEpoch = loadLastProcessedEpoch();

function connect() {
  const ws = new WebSocket(BOB_WS);

  ws.on('open', () => {
    const filter = { identity: DEPOSIT_IDENTITIES };
    // Resume where we stopped: replays everything missed while disconnected.
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
      // Must be idempotent: catch-up can replay events you already credited.
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

Two things this example is careful about, and your implementation should be too:

- **Idempotent crediting.** Catch-up replays events you may have already processed, so key each credit on
  something stable (`epoch` + `logId`) rather than crediting blindly.
- **Commit the cursor after the credit**, not before — otherwise a crash between the two silently loses a deposit.

> [!NOTE]
> `transfers` only emits QU transfers that actually settled, so no execution check is needed here — unlike
> `tickStream` below, which also surfaces transactions that never executed.

#### Full tick stream

The `transfers` subscription pushes only settled QU-transfer log events. If you would rather receive **every
transaction per tick together with its execution status**, use the `tickStream` subscription instead. For each
tick it returns the list of transactions — each with an `executed` boolean and the log range it produced
(`logIdFrom`, `logIdLength`) — alongside the tick's log events. Narrow it with `txFilters`
(`from` / `to` / `minAmount` / `inputType`) and catch up from a past tick with `startTick`:

```json
{
  "jsonrpc": "2.0",
  "method": "qubic_subscribe",
  "params": ["tickStream", { "txFilters": [{ "to": "YOUR_DEPOSIT_IDENTITY" }] }],
  "id": 1
}
```

> [!IMPORTANT]
> Unlike `transfers`, a `tickStream` transaction list also includes transactions that did **not** execute
> (`executed: false`). Always require `executed: true` before crediting a deposit — this is Bob's equivalent of
> the `moneyFlew` / `isExecuted` check in the [TX-based workflow](tx-based-workflow.md).

Trade-off: `tickStream` is more comprehensive (full transactions + logs + execution status in one stream, ideal
if you are effectively building your own indexer) but heavier than the specialized `transfers` filter. Full
parameters and payloads are in
[QUBIC_JSON_RPC.md → WebSocket Subscriptions](https://github.com/qubic/core-bob/blob/master/docs/QUBIC_JSON_RPC.md#websocket-subscriptions).

### Broadcasting transactions

A self-hosted Bob can broadcast signed transactions, so you do **not** need the public Live API to send
withdrawals. The flow mirrors the
[TX-based workflow](tx-based-workflow.md#creating-signing-sending-and-verifying-a-transaction) (including the
Qutil / Send-Many path) with two calls repointed at Bob:

**1. Get the current tick** to set a target tick in the future (the Live API's `tick-info` equivalent):

```bash
curl -s -X POST http://<your-bob-host>:40420/qubic \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"qubic_getTickNumber","params":[],"id":1}'
```

**2. Build and sign** the transaction with `go-node-connector` exactly as in the TX-based workflow — unchanged,
since signing is network format, not RPC. Set `targetTick` to the value from step 1 plus a buffer: the TX-based
workflow [recommends 15–20 ticks ahead](tx-based-workflow.md#respect-status-information) (its examples use `+15`)
so the transaction reaches the network before that tick passes. This is a network rule, unchanged by Bob.

**3. Serialize the signed transaction to hex.** This is the one real difference from the public Live API. The
workflow's `lsc.BroadcastTransaction(tx)` serializes and base64-encodes the transaction for you; Bob instead
takes the raw signed bytes as a **hex** string, so you serialize it yourself:

```go
tx, _ = signer.SignTx(tx)             // signed transaction (from step 2)
packet, _ := tx.MarshallBinary()      // full signed bytes, signature included
data := hex.EncodeToString(packet)    // Bob wants HEX
```

> [!WARNING]
> Bob's `data` field is **hex**, not base64. `go-node-connector`'s `EncodeToBase64()` (what the public Live API
> uses) produces base64 — do **not** send that to Bob. Use `MarshallBinary()` + `hex.EncodeToString`.

**4. Broadcast** the hex payload (optional `0x` prefix, even length) to your node:

```bash
curl -s -X POST http://<your-bob-host>:40420/broadcastTransaction \
  -H "Content-Type: application/json" \
  -d '{"data":"<signed-transaction-hex>"}'
```

The response is a `BroadcastResult` (`success`, `txHash`, `error`). The equivalent JSON-RPC method is
`qubic_broadcastTransaction`. See the [REST API reference](https://github.com/qubic/core-bob/blob/master/docs/REST_API.md)
for the full endpoint contract.

> [!NOTE]
> The same one-concurrent-transaction-per-source-address and target-tick rules from the
> [workflow guidelines](tx-based-workflow.md#qubic-workflow-guidelines) still apply — they are network rules, not
> RPC-specific. The tick you target must come from **step 1 (Bob's live view)**, not from
> `GET /status`'s `currentIndexingTick`, which is how far Bob has *indexed* and lags the network.

### Verifying transactions

After broadcasting, verify execution and retrieve the resulting log events via Bob's REST API
(`GET /tx/{tx_hash}` → `isExecuted`, `logIdFrom`, `logIdTo`, then `GET /log/{epoch}/{from}/{to}`). This flow is
documented in [DEAL_WITH_TX.MD](https://github.com/qubic/core-bob/blob/master/docs/DEAL_WITH_TX.MD).

### Useful REST endpoints

WebSocket streams are best for *reacting* to new activity; the REST API is best for *asking* Bob a direct
question — health checks, balance lookups, one-off verification, and pulling historical ranges for reconciliation
without replaying a stream. The full contract is in
[REST_API.md](https://github.com/qubic/core-bob/blob/master/docs/REST_API.md); the endpoints most relevant to an
exchange:

| Endpoint | Purpose |
| -------- | ------- |
| `GET /status` | Node sync/health (`isSyncing`, `progress`, indexing vs network tick). Use for monitoring and to confirm catch-up. |
| `GET /epochinfo/{epoch}` | What tick range the node holds for an epoch, and whether it has that epoch at all (`found`). |
| `GET /balance/{identity}` | Current balance of an identity — useful for reconciliation and withdrawal pre-checks. |
| `GET /tx/{tx_hash}` | Verify a specific transaction (`isExecuted`, log range). |
| `GET /log/{epoch}/{from_id}/{to_id}` | Fetch the log events a transaction produced. |
| `POST /getQuTransfersForIdentity` | QU transfers for a deposit identity within a tick range — historical backfill / gap reconciliation. |
| `POST /getAssetTransfersForIdentity` | Asset transfers for an identity (by issuer + asset name) within a tick range. |
| `GET /asset/{identity}/{issuer}/{asset_name}/{manageSCIndex}` | Asset balance for an identity. |
| `POST /broadcastTransaction` | Send a signed transaction (see [above](#broadcasting-transactions)). |
| `POST /querySmartContract` | Query smart-contract state (e.g. for future use cases). |

> [!TIP]
> A common pattern is: subscribe over WebSocket for real-time deposit detection, and use
> `POST /getQuTransfersForIdentity` over a tick range to **reconcile / backfill** after any downtime, bounded by
> your [retention window](#data-retention).

---

## Open questions to resolve before publishing

> [!WARNING]
> This section is for the internal review of this draft and should be **removed before publishing**. Colleagues —
> please help complete these:

1. **Trusted peer nodes.** Bob requires at least one `trusted-node` / `p2p-node` to sync from. **Who provides
   these addresses to exchanges, and how?** Is there an official/maintained list of Qubic peer nodes exchanges
   should point to, do they get dedicated peers, or are they expected to run/obtain their own? This is the
   biggest blocker for the install/config sections being actionable.
2. **Non-seamless / network-reset transitions.** The seamless case is now documented (3-step workflow +
   `wait_at_epoch_end`, per the Bob author). Still to confirm: does a *non-seamless* event (network reset/restart)
   need anything beyond the same auto-restart, e.g. a deliberate storage wipe and re-sync?
3. **`spam-qu-threshold` default.** Is there a recommended value for exchanges, or should it stay `0`?
4. **Cold-start / full-resync duration.** Roughly how long does a from-scratch sync take? (The "caught up" signal
   is answered — `GET /status` reports `isSyncing` / `progress`.)
5. **Security hardening.** Any recommendations we should add — e.g. not exposing port `40420` publicly, keeping
   `enable-admin-endpoints: false`, network placement relative to the exchange's hot wallet?
6. **Sizing over time.** With the default `kvrocks_ttl` (14 days ≈ 2 epochs) growth is bounded — but what disk
   footprint should an exchange actually expect for that window, and how does it scale if they raise the TTL?
7. **`KVROCKS_ROCKSDB_TTL` is a no-op — to be fixed in core-bob, not worked around here.** The env var maps to
   `rocksdb.ttl`, which Kvrocks rejects as an invalid configuration key (`docker exec qubic-bob grep -i "not a
   valid configuration key" /app/logs/kvrocks.log`), so it silently does nothing. It is still present in
   `docker/kvrocks.conf`, `docker/entrypoint.sh`, and is still described in `DOCKER_ENV.md` as a "RocksDB-level
   secondary TTL safety net" that should sit slightly above `KVROCKS_TTL`. The Bob team has acknowledged this and
   intended to remove it; it has not landed yet.
   **Ask them to resolve it in core-bob** — whether that means removing the var outright or documenting it
   differently (e.g. if a working equivalent exists under another key) is their call.
   **Why this blocks us:** we link to `DOCKER_ENV.md`, so until it is corrected a reader can still be told the
   safety net exists. That assumption already brought an operator close to filling a disk. If it is not resolved
   before we publish, we should add a warning here rather than link readers to misleading guidance.
8. **Expose `epochTickQuality` in Bob (concrete feature request).** Of the `/v1/latest-stats` fields, Bitget
   confirmed they use exactly **one: `epochTickQuality`** — the rest is market data they source elsewhere. This is
   *not* market data; it is chain-derived (`epochTickQuality ≈ (ticksInEpoch − emptyTicks) / ticksInEpoch`).
   What Bob has today: the epoch's tick **range** via `epochinfo` (so *total* ticks is derivable as
   `endTick − initialTick + 1`) and *per-tick* empty status (via `GET /tick/{n}` or `tickStream` empty-tick
   notifications). What it does **not** have: any stored/exposed **empty-tick count** or quality aggregate —
   verified across the `EpochInfo` / `SyncStatus` structs and the indexer / processor / database code, not just the
   docs (zero hits). So deriving `epochTickQuality` today means scanning every tick or counting the whole stream —
   not a single query.
   **Why an exchange wants it:** `epochTickQuality` is a network-health / liveness gauge (share of produced vs.
   total ticks). Exchanges use it for risk gating (throttle or pause deposits/withdrawals when quality drops),
   confirmation policy (wait for more ticks during degraded periods), ops dashboards/alerting, and incident triage
   (network-degraded vs. own-bug).
   **Ask the Bob team to expose it directly** — a field on `GET /status` or `GET /epochinfo/{epoch}`, or the raw
   `ticksInEpoch` / `emptyTicks` counts so the exchange can compute it. Until then, exchanges must keep calling the
   public Stats API `/v1/latest-stats` for this single value even when everything else runs on their own Bob node.
9. **Bob docs lack a build → sign → broadcast example.** Bob's documentation only shows `broadcastTransaction`
   taking an already-signed hex payload; it never shows how to construct or sign a transaction, and does not
   mention `go-node-connector` (signing is out of Bob's scope by design). That leaves a gap for integrators — in
   particular the base64-vs-**hex** difference from the public Live API, which we bridge in
   [Broadcasting transactions](#broadcasting-transactions). Worth asking the Bob team to add a short end-to-end
   signing example (in Go and/or the JS/TS library) to the Bob repo so it is not documented only here.
10. **`getProcessedTickIntervals` multi-interval nuance.** The public endpoint can return *multiple* tick
   intervals for a single epoch (when the network restarts mid-epoch); Bob's `GET /epochinfo/{epoch}` reports one
   `initialTick` / `endTick` per epoch. Confirm with the Bob team whether `epochinfo` captures a mid-epoch
   restart (or whether an integration that depends on detecting split intervals would miss it).

---

The canonical documentation for Bob lives in the [Bob repository](https://github.com/qubic/core-bob) (linked
throughout this guide). See also the [Partner integration overview](README.md).
