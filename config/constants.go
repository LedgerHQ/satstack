package config

import "time"

// LedgerNanoSGenesis indicates the earliest possible date of a Ledger device
// that is currently supported by Ledger Live.
//
// Ledger Nano S was launched in 2016/06/01.
var LedgerNanoSGenesis, _ = time.Parse("2006/01/02", "2016/06/01")
