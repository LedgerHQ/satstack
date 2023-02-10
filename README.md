# Ledger SatStack

<img src="/share/logo.svg" align="right" alt="satstack logo" width="200">

</h1>

Ledger SatStack is a lightweight bridge to connect Ledger Live with your personal Bitcoin full node. It's designed to allow Ledger Live users use Bitcoin without compromising on privacy, or relying on Ledger's infrastructure.

<p>
  <img src="https://github.com/ledgerhq/satstack/workflows/Build/badge.svg" />
  <img src="https://img.shields.io/github/v/release/ledgerhq/satstack?include_prereleases" />
  <img src="https://img.shields.io/github/downloads/ledgerhq/satstack/total">
  <img src="https://img.shields.io/badge/Go-%3E%3D1.17-04ADD8.svg" />
</p>

<img src="share/screenshot.png" align="center" />

## Table of Contents

- [Background](#background)
- [Architecture](#architecture)
- [Requirements](#requirements)
- [Usage](#usage)
- [Misc](#misc)
- [In the Press](#in-the-press)
- [Community](#community)
- [Contributing](#contributing)

### Background

Running a full node is the only way you can use Bitcoin in a completely trustless way. A full node downloads the entire blockchain, and checks it against Bitcoin's consensus rules, and contributes to the decentralization and economic strength of Bitcoin. However, a far more compelling reason to run your own node is **privacy**. [...read more](https://en.bitcoin.it/wiki/Full_node).

Running a node can be difficult for some users, and has [associated costs](https://bitcoin.org/en/full-node#costs-and-warnings) in terms of network bandwidth and disk usage. This is why Live connects to Bitcoin nodes running on Ledger's infrastucture, wrapped around by indexer and explorer services to ensure fast queries. While security and privacy is core to Ledger, one can make a theoretical case that Ledger can spy on transaction details, or even censor certain addresses from using Ledger's services.

SatStack aims to render Ledger's infrastructure dispensable, by allowing users to connect Ledger Live with their personal Bitcoin full node.

### Architecture

Ledger SatStack is a standalone Go application, that acts as a bridge between the [Ledger Live](http://ledger.com/live) application and a Bitcoin Core full-node. It exposes a REST interface to the open-source JS library embedded by Live, and communicates to the Bitcoin node over RPC. It utilizes the transport layer and data-structures of [btcd](https://github.com/btcsuite/btcd).

<p align="center">
  <img src="/share/architecture.svg"/>
</p>

### Requirements

- Bitcoin Nano app **`2+`**
- Bitcoin Core **`0.22.0+`**
- Ledger Live (desktop) **`2.44.0+`** but don't go as far 2.53+ that breaks satstack! https://download.live.ledger.com/ to get the latest supported i.e. 2.52.0
- `txindex=1` in `bitcoin.conf` is not mandatory, but recommended. Pruned nodes are not currently supported.
- Wallet should **NOT** be disabled (attn. Raspiblitz users).

### Usage

#### Setup Ledger Live (recommended way but only works without Tor)

The easiest way of getting started is to use the dedicated setup flow directly on Ledger Live.
A detailed guide is available [here](https://support.ledger.com/hc/en-us/articles/360017551659).

#### Manual setup (for advanced users)

##### Retrieve descriptors from device

Simply follow these steps:

1. Plug in your Ledger device via USB.
2. Enter your PIN code on the device, and open the Bitcoin app.
3. Run the `scripts/getdescriptor` script, as shown below.

```bash
$ cd scripts
$ python3 -m venv venv  # ensure Python 3.8+
$ source venv/bin/activate
(venv) $ pip install -r requirements.txt
(venv) $ ./getdescriptor --scheme native_segwit --chain main --account 3
External: wpkh([b91fb6c1/84'/0'/3']xpub6D1gvTP...VeMLtH6/0/*)
Internal: wpkh([b91fb6c1/84'/0'/3']xpub6D1gvTP...VeMLtH6/1/*)
```

if you get an `unsupported hash type ripemd160` error, please see [this](https://stackoverflow.com/questions/72409563/unsupported-hash-type-ripemd160-with-hashlib-in-python)

##### Create configuration file

Create a config file **`lss.json`** in your home directory.
You can use [this](https://github.com/ledgerhq/satstack/blob/master/lss.mainnet.json) sample config file as a template.

Add `"torproxy": "socks5://127.0.0.1:9050",` to connect to a Tor client running locally so that satstack can reach a full node behind Tor.
Replace the `rpcurl` with the .onion address of your node.

###### Optional account fields

- **`depth`**: override the number of addresses to derive and import in the Bitcoin wallet. Defaults to `1000`.
- **`birthday`**: set the earliest known creation date (`YYYY/MM/DD` format), for faster account import.
  Defaults to `2013/09/10` ([BIP0039](https://github.com/bitcoin/bips/blob/master/bip-0039.mediawiki) proposal date).
  Refer to the table below for a list of safe wallet birthdays to choose from.

  | Event                                                    | Date (YYYY/MM/DD)    |
  | -------------------------------------------------------- | -------------------- |
  | BIP0039 proposal created                                 | 2013/09/10 (default) |
  | First ever BIP39 compatible Ledger device (Nano) shipped | 2014/11/24           |
  | First ever Ledger Nano S shipped                         | 2016/07/28           |

##### Launch Bitcoin full node

Make sure you've read the [requirements](#requirements) first, and that your node is configured properly.
Here's the recommended configuration for `bitcoin.conf`:

It is not recommended to use rpcuser/rpcpassword in the bitcoin.conf of your full node use rpcauth instead.
If you still want to use it make sure the following settings are in your bitcoin.conf file:

```
# Enable RPC server
server=1

# Enable indexes
txindex=1
blockfilterindex=1

# Set RPC credentials
rpcuser=<user>
rpcpassword=<password>
```

If you want to use the newest security standard recommended by the core-devs then read the following paragraph.
Get the rpcauth.py script from the bitcoin repo and create a new user for satstack.

```
wget https://raw.githubusercontent.com/bitcoin/bitcoin/master/share/rpcauth/rpcauth.py

python3 rpcauth.py satstack

Example Output:
rpcauth=satstack:a14191e6892facf70686a397b126423$ddd6f7480817bd6f8083a2e07e24b93c4d74e667f3a001df26c5dd0ef5eafd0d
Your password:
VX3z87LBVc_X7NBLABLABLABLA
```

Copy the `rpcauth=` into your bitcoin.conf
Note down the password and use them in your lss.json. There you will use the credentials

```
In your lss.json:

"rpcuser": "satstack",
"rpcpassword": "VX3z87LBVc_X7NBLABLABLABLA"
```

```config
# Enable RPC server
server=1

# Enable indexes
txindex=1
blockfilterindex=1

# Set RPC credentials
# Example Auth, replace with your own.
# see https://bitcoin.stackexchange.com/questions/46782/rpc-cookie-authentication why we are not using normal rpcuser/password
rpcauth=satstack:a14191e6892facf70686a397b126423$ddd6f7480817bd6f8083a2e07e24b93c4d74e667f3a001df26c5dd0ef5eafd0d
```

Then launch `bitcoind` like this:

```bash
bitcoind
```

##### Launch SatStack

Pre-built binaries are available for download on the [releases](https://github.com/ledgerhq/satstack/releases)
page (Linux, Windows, MacOS). Extract the tarball, and launch it as:

```sh
./lss
```
There are some cmd handles which can be useful when configuring the setup the first time. List them with

`./lss -h` or `./lss --help`

When setting up a new wallet, the wallet is synced form the birthday date or your custom date set in `lss.json`
When the initial sync sucessfully completes, satstack saves a file called `lss_rescan.json` at the exact location
where the lss.json is stored. This file includes the latest blockheight your wallet was synced to, this allows 
satstack on every restart to not rescan the whole wallet again but only rescan the difference between the current
blockheight. You can also change the latest blockheight manually in the file which helps you to set the rescan delta
manually. This file is only created when an initial wallet sync was successful. Removing the file will lead satstack
to rescan the complete wallet again when starting up.

If you want to build `lss` yourself, just do the following:

(make sure you have [mage](https://magefile.org) installed first)

```sh
mage release  # or "mage build" for a development build
```

On startup, SatStack will wait for the Bitcoin node to be fully synced,
and import your accounts. This can take a while.

##### Launch Ledger Live Desktop

```sh
# environment variables `EXPLORER` and `EXPLORER_SATSTACK` should point at the address
# where `lss` is listening (can be a differnet computer/server)
$ EXPLORER=http://127.0.0.1:20000 <Ledger Live executable>
```

### Misc

If you get `error=failed to load wallet: -4: Wallet file verification failed. SQLiteDatabase: Unable to obtain an exclusive lock on the database, is it being used by another bitcoind?` maybe this is because you have bitcoind windows opened, if this is the case, please try closing them and restart lss.

### In the press

| Title                                                                                                                                                                                                               |                      Source                      |
| :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | :----------------------------------------------: |
| 🇬🇧 [Personal sovereignty with Ledger SatStack](https://blog.ledger.com/satstack)                                                                                                                                    |    [blog.ledger.com](https://blog.ledger.com)    |
| 🇫🇷 [Ledger SatStack: un pont entre Bitcoin Core et votre Ledger Wallet](https://bitcoin.fr/ledger-sat-stack-un-pont-entre-bitcoin-core-et-votre-ledger-wallet/)                                                     |         [bitcoin.fr](https://bitcoin.fr)         |
| 🇫🇷 [Votre propre coffre-fort à bitcoins… inviolable – Ledger annonce l’arrivée des full nodes Bitcoin](https://journalducoin.com/actualites/coffre-fort-bitcoins-inviolable-ledger-annonce-noeuds-complets-bitcoin) |   [Journal du Coin](https://journalducoin.com)   |
| 🇫🇷 [Il est désormais possible d’exécuter un full node Bitcoin sur Ledger Live](https://fr.beincrypto.com/technologie/5770/full-node-bitcoin-ledger-live)                                                            |     [beincrypto.com](https://beincrypto.com)     |
| 🇪🇸 [Ledger Live será compatible con nodos propios de Bitcoin](https://www.criptonoticias.com/tecnologia/ledger-live-sera-compatible-nodos-propios-bitcoin)                                                          | [CriptoNoticias](https://www.criptonoticias.com) |
| 🇬🇧 [Bitcoin Tech Talk #218: Curing Monetary Stockholm Syndrome](https://jimmysong.substack.com/p/curing-monetary-stockholm-syndrome) (mention)                                                                      |   [Jimmy Song](https://jimmysong.substack.com)   |

### Community

For feedback or support, please tag [@adrien_lacombe](https://twitter.com/adrien_lacombe) and [@Ledger](https://twitter.com/Ledger) on Twitter. To report any bugs related to full node on Ledger Live, you can create issues on this repository.

### Contributing

Contributions in the form of code improvements, documentation, tutorials, and feedback are most welcome.

For contributions to the code, we recommend [these guidelines](https://github.com/btcsuite/btcd/blob/master/docs/code_contribution_guidelines.md).

#### Call for Cowsay contributions

On startup, satstack will display a message about Bitcoin, randomly picked from a curated collection of interesting quotes, facts, email excerpts, etc. You are welcome to contribute by creating a [pull request](https://docs.github.com/en/free-pro-team@latest/github/collaborating-with-issues-and-pull-requests/creating-a-pull-request) modifying [this file](fortunes/db.go) (includes guidelines for editing the file). Here's an example of how it is rendered:

<img src="/share/cowsay.png" width="300">

##### Cowsay ideas

- Extracts from [The Complete Satoshi](https://satoshi.nakamotoinstitute.org) by the [Nakamoto Institute](https://nakamotoinstitute.org).
- Quotes by Satoshi, Hal Finney.
- Excerpts from bitcointalk.org, email lists, etc.
- Public criticisms of Bitcoin by famous people, media, etc. [Bitcoin Obituaries](https://99bitcoins.com/bitcoin-obituaries) is a great source.

Please mention the source when you make a contribution, so we can attribute the original author(s) and include a copy of the license if required.
