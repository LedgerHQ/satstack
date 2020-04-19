package utils

import "time"

// ParseUnixTimestamp converts a UNIX timestamp in seconds, and returns a
// string represention of the timestamp in RFC3339 format.
func ParseUnixTimestamp(timestamp int64) string {
	return time.Unix(timestamp, 0).UTC().Format(time.RFC3339)
}

// ParseSatoshi converts a float64 bitcoin value to satoshis.
// Named after ParseInt function.
func ParseSatoshi(value float64) int64 {
	return int64(value * 100000000)
}
