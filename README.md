# Ledger Sats Stack

<img src="/docs/logo.png" align="right" 
    alt="Legder Sats Stack logo by Anton Lovchikov" width="150">
     
<!--<a href="https://golang.org"><img height="100" src="https://i.pinimg.com/originals/12/5c/e0/125ce0baff3271761ca61843eccf7985.jpg" alt="Gopher" /></a>-->

</h1>

Ledger Sats Stack is a lightweight bridge to connect Ledger Live with your personal Bitcoin full node. It's designed to allow Ledger Live users use Bitcoin without compromising on privacy, or relying on Ledger's infrastructure.

<p>
  <img src="https://github.com/onyb/ledger-sats-stack/workflows/Build/badge.svg" />
  <img src="https://github.com/onyb/ledger-sats-stack/workflows/reviewdog/badge.svg" />
  <img src="https://github.com/onyb/ledger-sats-stack/workflows/Integration%20tests/badge.svg" />
  <img src="https://github.com/onyb/ledger-sats-stack/workflows/Regression%20tests/badge.svg" />
  <img src="https://img.shields.io/badge/Go-%3E%3D1.13-orange.svg" />
</p>

# Table of Contents

- [Background](#background)
- [Architecture](#architecture)
- [Requirements](#requirements)
- [Usage](#usage)
- [Contribute](#contribute)

## Background

Running a full node is the only way you can use Bitcoin in a completely trustless way. A full node downloads the entire blockchain, and checks it against Bitcoin's consensus rules, and contributes to the decentralization and economic strength of Bitcoin. However, a far more compelling reason to run your own node is **privacy**. [...read more](https://en.bitcoin.it/wiki/Full_node).

Running a node can be difficult for some users, and has [associated costs](https://bitcoin.org/en/full-node#costs-and-warnings) in terms of network bandwidth and disk usage. This is why Live connects to Bitcoin nodes running on Ledger's infrastucture, wrapped around by indexer and explorer services to ensure fast queries. While security and privacy is core to Ledger, one can make a theoretical case that Ledger can spy on transaction details, or even censor certain addresses from using Ledger's services.

Sats Stack aims to render Ledger's infrastructure dispensable, by allowing users to connect Ledger Live with their personal Bitcoin full node.


## Architecture

Ledger Sats Stack is a standalone Go application, that acts as a bridge between the [Ledger Live](http://ledger.com/live) application and a Bitcoin Core full-node. It exposes a REST interface to the open-source C++ library [libcore](https://github.com/LedgerHQ/lib-ledger-core), embedded by Live, and communicates to the Bitcoin node over RPC. It utilizes the transport layer and data-structures of [btcd](https://github.com/btcsuite/btcd).

<p align="center">
  <img src="/docs/architecture.png"/>
</p>

## Requirements

- Bitcoin Core 0.18.0+
  * Must have RPC enabled. Make sure you have this in `bitcoin.conf`.
    ```
    server=1
    rpcuser=<user>
    rpcpassword=<password>
    ```
  * Must have `txindex` enabled. Make sure you have this in `bitcoin.conf`.
    ```
    txindex=1
    ```
    
    Will be optional after [**`#15`**](https://github.com/onyb/ledger-sats-stack/issues/15).
- Ledger Live \<insert version\>

## Usage

##### Create configuration file

Gather the information of accounts you want to track from Ledger Live, and add them to a `~/.sats.json` file in the following format:

```json
[
  {
    "xpub": "xpub...",
    "index": 0,
    "type": "segwit"
  },
  {
    "xpub": "xpub...",
    "index": 1,
    "type": "segwit"
  }
]
```

##### Launch Bitcoin full node

Make sure you meet the [requirements](#requirements) first, then launch Bitcoin like this:

```
$ bitcoind -datadir=~/.bitcoin
```

It's not important to have the node completely synced before proceeding to the next step.

##### Launch Sats Stack

```sh
$ git clone https://github.com/onyb/ledger-sats-stack
$ cd ledger-sats-stack
$ BITCOIND_RPC_HOST=localhost:8332 BITCOIND_RPC_USER=<user> BITCOIND_RPC_PASSWORD=<password> make dev
```

**Note:** The RPC port is `18332` for `test` chain.

##### Launch Ledger Live Desktop

```sh
$ git clone https://github.com/ledgerhq/ledger-live-desktop
$ cd ledger-live-desktop
$ yarn
$ EXPERIMENTAL_EXPLORERS=1 EXPLORER=http://0.0.0.0:20000 yarn start
```

## Contribute

Contributions in the form of code improvements, documentation, tutorials, and feedback are most welcome. If you want to help out, please see [CONTRIBUTING.md](/).
