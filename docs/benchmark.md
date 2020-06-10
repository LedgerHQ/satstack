## Methodology

Currently, it is not possible to run benchmarking tests on the CI, since it
requires a full node with the account descriptors imported in the bitcoind
wallet. I used [this script](tests/integration/sync.sh) for the benchmarks,
but the process is still very manual. If you have ideas on improving this,
let's talk!


⚠️ These benchmarks should not be taken at face value. The synchronization
time also has a lot to do with the number of UTXOs, which is not mentioned
below.

There are a lot of ways to improve the performance, although **RPC Batch**
and/or **Goroutines** are probably going to have the biggest impact.


#### With `txindex=1`


| xpub                  | Operations | Ledger Explorer | Ledger Sats Stack |
| :--------------------:|-----------:|----------------:|------------------:|
| `tpubDDTG...FeF6mSjZ` | 2          | 4.40s           | 9.19s             |
| `tpubDDbk...auTWxMLC` | 6          | 4.11s           | 10.66s            |
| `tpubDDTG...H9UNPw4s` | 8          | 4.84s           | 8.02s             |
| `tpubDDAt...tHEHYUnt` | 18         | 3.97s           | 11.04s            |
| `tpubDCHC...9mFJejwC` | 30         | 4.00s           | 10.83s            |
| `tpubDCkv...kFk9DBpZ` | 36         | 6.69s           | 11.00s            |
| `tpubDCuo...wrhHqhsW` | 928        | 16.57s          | 57.74s            |
