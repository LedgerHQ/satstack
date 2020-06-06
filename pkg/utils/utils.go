package utils

import (
	"strings"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
)

// ParseUnixTimestamp converts a UNIX timestamp in seconds, and returns a
// string representing of the timestamp in RFC3339 format.
func ParseUnixTimestamp(timestamp int64) string {
	return time.Unix(timestamp, 0).UTC().Format(time.RFC3339)
}

// ParseSatoshi converts a float64 bitcoin value to satoshis.
// Named after ParseInt function.
func ParseSatoshi(value float64) btcutil.Amount {
	// Convert BTC value to satoshi without losing precision.
	amount, err := btcutil.NewAmount(value)
	if err != nil {
		// TODO: Log an error here
		return -1
	}
	return amount
}

func ParseChainHash(hash string) (*chainhash.Hash, error) {
	return chainhash.NewHashFromStr(strings.TrimLeft(hash, "0x"))
}

// Contains is a helper function to check if a string
// exists in a slice of strings.
func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
