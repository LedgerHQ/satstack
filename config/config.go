package config

import (
	"ledger-sats-stack/utils"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// Account struct models the configuration of an account on Ledger Live.
//
// Fields marked as (?) are optional.
type Account struct {
	Descriptor     *string `json:"descriptor"`     // output descriptor for the account
	XPub           *string `json:"xpub"`           // xPub of the account
	Index          *int    `json:"index"`          // the account index
	DerivationMode *string `json:"derivationMode"` // standard, segwit, or native_segwit
	DerivationPath *string `json:"derivationPath"` // (?) Will override libcore defaults
	Birthday       *date   `json:"birthday"`       // (?) Earliest known creation date (YYYY/MM/DD)
}

// Configuration is a struct to model the JSON configuration
// of the project, stored in ~/.lss.json file.
//
// Fields marked as (?) are optional.
type Configuration struct {
	RPCURL      *string   `json:"rpcURL"`
	RPCUser     *string   `json:"rpcUser"`
	RPCPassword *string   `json:"rpcPassword"`
	RPCTLS      bool      `json:"rpcTLS"`
	Accounts    []Account `json:"accounts"`
	Depth       *int      `json:"depth"` // (?) Number of addresses to import
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
		if account.Descriptor == nil {
			validateStringField("xpub", account.XPub)
			validateIntField("index", account.Index)
			validateStringField("derivationMode", account.DerivationMode)

			validDerivationModes := []string{"standard", "segwit", "native_segwit"}
			if !utils.Contains(validDerivationModes, *account.DerivationMode) {
				log.WithFields(log.Fields{
					"derivationMode": *account.DerivationMode,
				}).Fatal("Invalid value for field")
			}
		}
	}
}

func validateStringField(key string, value *string) {
	if value == nil {
		log.WithFields(log.Fields{
			key: value,
		}).Fatal("Missing configuration")
	}
}

func validateIntField(key string, value *int) {
	if value == nil {
		log.WithFields(log.Fields{
			key: value,
		}).Fatal("Missing configuration")
	}
}
