# TX Based Exchange integration

## Management Summary
This document describes the Qubic RPC infrastructure and how it can be used to integrate Qubic into an exchange environment.

## About qubic
* Launched: 13.4.2022 (mainnet)
* Fair Launch: no VC or team allocations
* Circulating Supply: 71.4trn (trilion)
* Weekly Emission: 1trn
* Hard Cap/Max Supply: 1000trn (will most likely never be reached due to burn mechanisms)

### Available to buy
* [SafeTrade](https://safe.trade/)
* [Tradeogre](https://tradeogre.com/markets)
* [Seven Seas Exchange](https://www.sevenseas.exchange/)
* [Bitkonan](https://www.bitkonan.com)

### Trackers
* [Coinmarketcap](https://coinmarketcap.com/currencies/qubic/)
* [Coingecko](https://www.coingecko.com/en/coins/qubic-network)
* [Delta mobile app](http://delta.app)
* [Livecoinwatch](https://www.livecoinwatch.com/price/QUBIC-QUBIC)
* [Coinpaprika](https://coinpaprika.com/coin/qubic-qubic/)


## Exchange Related Information
* Project Name: Qubic
* Ticker: QUBIC 
* Currency: QU
* Brand assets (logo, icon) for [download](https://drive.google.com/file/d/13XcdR7nWMTNVRL5J5FYZyc-SnMMAYN3w/view)
* Decimals: The smallest possible unit is 1 QU (like 1 Satoshi in Bitcoin) and it trades at a price of around 0.000006879 USD right now.

### Links
* Github: [https://github.com/qubic](https://github.com/qubic)
* X/Twitter: [https://x.com/_Qubic_](https://x.com/_Qubic_) 
* Website: [https://qubic.org/](https://qubic.org/)
* Discord: [https://discord.com/invite/qubic](https://discord.com/invite/qubic)
* Telegram: [https://t.me/qubic_network](https://t.me/qubic_network)
* Youtube: [https://www.youtube.com/@_qubic_](https://www.youtube.com/@_qubic_)
* Explorer: [https://app.qubic.li/network/explorer](https://app.qubic.li/network/explorer) 
* Medium, Reddit, Facebook, Instagram: n/a

### Tech Integration Basics
* Base URL for production environment: `https://ex.qubic.org/v1`
* Base URL for test environment (points to test infrastructure which accesses the production chain): `https://testapi.qubic.org/v1`
* Access to the production environment is IP restricted.
* Individual rate limits can be set by source ip address.
* The HTTP API is open and doesn’t require any special authentication.
* All infrastructure is protected by Cloudflare

> [!IMPORTANT]
> Use `https://testapi.qubic.org/v1` for testing and let us know your IP addresses in order to access `https://ex.qubic.org/v1`

### Tech FAQ
* **Is there a need to run an own node?** \
No, we suggest using our infrastructure as described in this document for easy onboarding. But you can also run your own node. Let’s discuss.
* **Does qubic support memo for wallets?** \
No, it doesn’t.
* **Where do I get test tokens?** \
Just get in touch, we’re happy to send some. \


# General Qubic Terms
This section is about terms needed to understand Qubic. It explains the dos and don’ts.

## QU (Qubics)
QU is the native currency of the Qubic network.

## Seed and ID
The **seed** is the equivalent to the private key in Bitcoin/Ethereum. As seed is a string of 55 low case characters [a-z]. Example: `lyvborkwdxwnghiohudjgrmvdadbecyvjrlrtqsyajpeajkgaxbohky`

The **ID** is the equivalent of the public key, it is derived from the seed and has a length on 70 large cap characters [A-Z]. Example: `RCZDMCEENYZPHGEUTIIOTWIKTNSCGWWQGKMOBNIVDCDTSYSBTMKQWRQECPHM`

## Epoch
An Epoch in Qubic is one week. It starts every Wednesday at 12 UTC and ends one week later before the new epoch starts.
Epochs are important because every epoch change gives the node operators the chance to update their nodes. Currently every week is expected to have one core update.
During epoch change ~11:00 to ~13:00 UTC the general network availability is deprecated. This means that asking for current balance may take a bit longer than normal.
Using the Qubic RPC interface you will have responsiveness also during epoch transition because all important data is cached.

> [!IMPORTANT]
> We suggest <strong>not</strong> to  send out transactions during the transition phase. Every Wednesday 11:45 UTC and 12:45 UTC.

## Computor
A Computor is one of 676 validators in Qubic.
Computors take on a multitude of tasks, among them are validation of ticks, execution of smart contracts, and validation of transactions. 

## Quorum
The Quorum (derived from a paper of [Nick Szabo](https://www.fon.hum.uva.nl/rob/Courses/InformationInSpeech/CDROM/Literature/LOTwinterschool2006/szabo.best.vwh.net/quorum.html), see also: [Quorum in distributed systems on Wikipedia](https://en.wikipedia.org/wiki/Quorum_(distributed_computing)). ) in Qubic is defined by 451 of 676 Computors.
The Quorum is needed to finalize Ticks or to make decisions. Only if 451 out of 676 Computors agree to a common state the network can proceed.
The Quorum is also needed to make network related decisions. This can be achieved by proposals and votes.

## Tick
A Tick in Qubic is a block for other blockchains. A Tick is defined by its TickData which is issued by the responsible Computor. A tick usually takes between two and eight seconds.
Every Tick can be non-empty or empty. Only non-empty ticks are considered as good and will include and execute transactions.
As soon as a tick has passed, the transactions (and everything else) included cannot be changed or reverted anymore (true finality).

> [!IMPORTANT]
> If a Tick is empty, transactions from this Tick are not included into blockchain and therefore not executed.

> [!IMPORTANT]
> Working with Ticks is the most important concept to understand in qubic. Integration will need to monitor ticks and act according to outcome (ie. re-issue a transaction or issue a new one - more on this later in this document)

## TickData
TickData is the definition of a Tick. It contains the cryptographically signed information what transactions must be processed in this Tick.
It also contains the timestamp of the Tick and other block information needed to determine if a Tick is valid or not.

## Transaction
A transaction in Qubic is a transport vehicle for multiple purposes.
The main purposes are:
1. Sending QU from one address to another
2. Sending Smart Contract commands

A transaction in Qubic needs to have a `Target Tick` when it should be executed. The network cannot decide this. It is good practice to add 5 Ticks of offset to the `latest Tick` (current block height).

> [!IMPORTANT]
>A Qubic transaction must have a `Target Tick`. The `Target Tick` can be chosen from `latest Tick` + 5

Due to Qubics architecture from one source address can **only exist one concurrent transaction** in the network.
Network is in tick 5. If Alice has one tx in the network for future tick 10 and then sends another one for tick 11 it would overwrite the first one. So one has to wait until tick 10 has passed. If Alice sends a tx for tick 9, nothing would change as the TX with the higher tick set will stay in the network.

> [!IMPORTANT]
> Only one concurrent TX per sending ID can exist. Don’t send a new transaction from the same source address until the tx `Target Tick` has been passed. This is important to consider for withdrawals: a hot wallet can only have one transaction, everything else needs to be queued. We discuss an approach ("SendMany") which allows for more bandwidth further down.

Transactions can be created and signed locally/offline. Only the signed transaction must be sent to the network.

# The Integration for Exchanges
The [tx based use case](tx-based-use-case.md) showcases the integration of qubic for exchanges: deposits and withdrawals based on the RPC services provided.


