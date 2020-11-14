## Methodology

Currently, it is not possible to run benchmarking tests on the CI, since it
requires a full node with the account descriptors imported in the bitcoind
wallet. I used [this script](tests/integration/sync.sh) for the benchmarks,
but the process is still very manual. If you have ideas on improving this,
let's talk!


⚠️ These benchmarks should not be taken at face value. The synchronization
time also has a lot to do with the number of UTXOs, which is not reflected
below.

There are a lot of ways to improve the performance, although **RPC Batch**
and/or **Goroutines** are probably going to have the biggest impact.

### Synchronization

##### Legend

* **Ops** = Number of operations (`libcore` abstraction for transactions)
* **LBE** = Ledger Blockchain Explorer
* **LSS** = Ledger SatStack
* **reset** = reset libcore DB and sync from scratch


#### Latest benchmark

##### With `txindex=0` + Native TX Parsing

| xpub                  | Ops | LBE (reset) | LSS (reset) | LBE         | LSS         |
|:---------------------:|----:|------------:|------------:|------------:|------------:|
| `tpubDDTG...FeF6mSjZ` | 5   | 7.04s       | 9.21s       | 3.08s       | 6.33s       |
| `tpubDDbk...auTWxMLC` | 15  | 5.30s       | 8.49s       | 3.05s       | 6.13s       |
| `tpubDDTG...H9UNPw4s` | 13  | 8.16s       | 12.73s      | 2.98s       | 8.68s       |
| `tpubDDAt...tHEHYUnt` | 28  | 6.75s       | 9.92s       | 3.16s       | 6.36s       |
| `tpubDCHC...9mFJejwC` | 40  | 5.41s       | 8.62s       | 3.12s       | 6.46s       |
| `tpubDCkv...kFk9DBpZ` | 44  | 5.50s       | 8.79s       | 3.16s       | 6.69s       |
| `tpubDCuo...wrhHqhsW` | 934 | 30.36s      | 42.89s      | 11.20s      | 33.50s      |


#### 🗄️ Benchmarks archived for posterity

##### With `txindex=0` ([`ecb0fc0`](https://github.com/onyb/sat-stack/tree/ecb0fc0))

| xpub                  | Ops | LBE (reset) | LSS (reset) | LBE         | LSS         |
|:---------------------:|----:|------------:|------------:|------------:|------------:|
| `tpubDDTG...FeF6mSjZ` | 5   | 9.67s       | 12.20s      | 3.31s       | 10.37s      |
| `tpubDDbk...auTWxMLC` | 15  | 5.38s       | 12.85s      | 3.04s       | 9.89s       |
| `tpubDDTG...H9UNPw4s` | 13  | 8.36s       | 18.52s      | 3.07s       | 18.32s      |
| `tpubDDAt...tHEHYUnt` | 28  | 6.35s       | 13.70s      | 2.97s       | 12.83s      |
| `tpubDCHC...9mFJejwC` | 40  | 6.42s       | 11.90s      | 5.83s       | 10.82s      |
| `tpubDCkv...kFk9DBpZ` | 44  | 5.53s       | 14.69s      | 3.43s       | 10.65s      |
| `tpubDCuo...wrhHqhsW` | 934 | 31.55s      | 76.51s      | 11.24s      | 67.87s      |


##### With `txindex=1` + Bus Cache enabled ([`ecb0fc0`](https://github.com/onyb/sat-stack/tree/ecb0fc0))


| xpub                  | Ops | LBE (reset) | LSS (reset) | LBE         | LSS         |
|:---------------------:|----:|------------:|------------:|------------:|------------:|
| `tpubDDTG...FeF6mSjZ` | 2   | 6.54s       | 8.70s       | 3.78s       | 5.04s       |
| `tpubDDbk...auTWxMLC` | 9   | 5.56s       | 7.90s       | 3.58s       | 4.18s       |
| `tpubDDTG...H9UNPw4s` | 10  | 6.90s       | 8.41s       | 3.68s       | 4.20s       |
| `tpubDDAt...tHEHYUnt` | 18  | 6.67s       | 8.87s       | 3.57s       | 5.31s       |
| `tpubDCHC...9mFJejwC` | 30  | 6.19s       | 8.21s       | 3.68s       | 4.67s       |
| `tpubDCkv...kFk9DBpZ` | 36  | 6.15s       | 8.17s       | 3.67s       | 4.99s       |
| `tpubDCuo...wrhHqhsW` | 928 | 33.13s      | 41.50s      | 13.04s      | 15.99s      |


##### With `txindex=1` + Bus Cache disabled  ([`4cbae0d`](https://github.com/ledgerhq/satstack/tree/4cbae0d))


| xpub                  | Ops | LBE (reset) | LSS (reset) | LBE         | LSS         |
|:---------------------:|----:|------------:|------------:|------------:|------------:|
| `tpubDDTG...FeF6mSjZ` | 2   | 6.59s       | 12.44s      | 3.58s       | 6.81s       |
| `tpubDDbk...auTWxMLC` | 9   | 5.76s       | 11.91s      | 3.68s       | 6.23s       |
| `tpubDDTG...H9UNPw4s` | 10  | 6.98s       | 10.43s      | 3.67s       | 4.94s       |
| `tpubDDAt...tHEHYUnt` | 18  | 6.81s       | 12.08s      | 3.56s       | 7.27s       |
| `tpubDCHC...9mFJejwC` | 30  | 5.96s       | 11.61s      | 3.66s       | 6.88s       |
| `tpubDCkv...kFk9DBpZ` | 36  | 6.21s       | 11.47s      | 3.88s       | 6.92s       |
| `tpubDCuo...wrhHqhsW` | 928 | 32.76s      | 66.40s      | 13.36s      | 24.83       |
