package config

import "time"

// BIP0039Genesis indicates the earliest date of a BIP39 seed that a Ledger
// device could possibly have.
var BIP0039Genesis, _ = time.Parse("2006/01/02", "2013/09/10")
