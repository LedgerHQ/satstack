package config

import (
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// Account struct models the configuration of an account on Ledger Live.
//
// Fields marked as (?) are optional.
type Account struct {
	Descriptor *string `json:"descriptor"` // output descriptor for the account
	Depth      *int    `json:"depth"`      // (?) Number of addresses to import
	Birthday   *date   `json:"birthday"`   // (?) Earliest known creation date (YYYY/MM/DD)
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

// Validate checks for the validity of the JSON configuration loaded in
// Configuration struct.
//
// It does not mutate the configuration values itself, but logs FATAL errors
// in case of invalid configuration.
func (c Configuration) Validate() {
	validateStringField("rpcURL", c.RPCURL)
	validateStringField("rpcUser", c.RPCUser)
	validateStringField("rpcPassword", c.RPCPassword)

	if c.Accounts == nil {
		log.WithFields(log.Fields{
			"accounts": c.Accounts,
		}).Fatal("Config validation failed")
	}

	for _, account := range c.Accounts {
		validateStringField("descriptor", account.Descriptor)

		// TODO: Add validation for birthday field
	}
}

func validateStringField(key string, value *string) {
	if value == nil {
		log.WithFields(log.Fields{
			key: value,
		}).Fatal("Missing configuration")
	}
}
