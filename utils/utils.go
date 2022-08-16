package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcjson"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// ParseUnixTimestamp converts a UNIX timestamp in seconds, and returns a
// string representing of the timestamp in RFC3339 format.
func ParseUnixTimestamp(timestamp int64) string {
	return time.Unix(timestamp, 0).UTC().Format(time.RFC3339)
}

// ParseRFC3339Timestamp converts a timestamp in RFC3339 format to a UNIX
// timestamp in seconds.
//
// It is the reverse of ParseUnixTimestamp
func ParseRFC3339Timestamp(timestamp string) (*int64, error) {
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return nil, err
	}

	tUnix := t.Unix()
	return &tUnix, nil
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

func ToStringPointer(value string) *string {
	return btcjson.String(value)
}

const (
	day  = time.Minute * 60 * 24
	year = 365 * day
)

// HumanizeDuration converts a time.Duration into something readable.
// Taken from: https://gist.github.com/harshavardhana/327e0577c4fed9211f65
//
// Example:
// 	HumanizeDuration(time.Duration())
func HumanizeDuration(d time.Duration) string {
	if d < day {
		return d.String()
	}

	var b strings.Builder

	if d >= year {
		years := d / year
		fmt.Fprintf(&b, "%dy", years)
		d -= years * year
	}

	days := d / day
	d -= days * day
	fmt.Fprintf(&b, "%dd%s", days, d)

	return b.String()
}
