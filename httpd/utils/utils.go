package utils

import "time"

// ParseUnixTimestamp converts a UNIX timestamp in seconds, and returns a
// string represention of the timestamp in RFC3339 format.
func ParseUnixTimestamp(timestamp int64) string {
	return time.Unix(timestamp, 0).UTC().Format(time.RFC3339)
}
