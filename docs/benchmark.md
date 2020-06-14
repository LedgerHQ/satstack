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

#### Legend

* **Ops** = Number of operations (`libcore` abstraction for transactions)
* **LBE** = Ledger Blockchain Explorer
* **LSS** = Ledger Sats Stack 
* **reset** = reset libcore DB and sync from scratch

### With `txindex=1`


| xpub                  | Ops | LBE (reset) | LSS (reset) | LBE         | LSS         |
| :--------------------:|----:|------------:|------------:|------------:|------------:|
| `tpubDDTG...FeF6mSjZ` | 2   | 6.59s       | 12.44s      | 3.58s       | 6.81s       |
| `tpubDDbk...auTWxMLC` | 9   | 5.76s       | 11.91s      | 3.68s       | 6.23s       |
| `tpubDDTG...H9UNPw4s` | 10  | 6.98s       | 10.43s      | 3.67s       | 4.94s       |
| `tpubDDAt...tHEHYUnt` | 18  | 6.81s       | 12.08s      | 3.56s       | 7.27s       |
| `tpubDCHC...9mFJejwC` | 30  | 5.96s       | 11.61s      | 3.66s       | 6.88s       |
| `tpubDCkv...kFk9DBpZ` | 36  | 6.21s       | 11.47s      | 3.88s       | 6.92s       |
| `tpubDCuo...wrhHqhsW` | 928 | 32.76s      | 66.40s      | 13.36s      | 24.83       |
