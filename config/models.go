package config

import (
	"strings"
	"time"
)

// Account struct models the configuration of an account on Ledger Live.
//
// Fields marked as (?) are optional.
type Account struct {
	External *string `json:"external"` // output descriptor at external path
	Internal *string `json:"internal"` // output descriptor at internal path
	Depth    *int    `json:"depth"`    // (?) Number of addresses to import
	Birthday *date   `json:"birthday"` // (?) Earliest known creation date (YYYY/MM/DD)
}

// Configuration is a struct to model the JSON configuration
// of the project, stored in ~/.lss.json file.
//
// Fields marked as (?) are optional.
type Configuration struct {
	RPCURL      *string   `json:"rpcurl"`
	RPCUser     *string   `json:"rpcuser"`
	RPCPassword *string   `json:"rpcpass"`
	NoTLS       bool      `json:"notls"`
	Accounts    []Account `json:"accounts"`
}

type date struct {
	time.Time
}

func (d *date) UnmarshalJSON(input []byte) error {
	strInput := string(input)
	strInput = strings.Trim(strInput, `"`)
	newTime, err := time.Parse("2006/01/02", strInput)
	if err != nil {
		return err
	}

	d.Time = newTime
	return nil
}
