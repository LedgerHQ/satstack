package config

import (
	"ledger-sats-stack/pkg/utils"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// Account struct models the configuration of an account on Ledger Live.
// Fields marked as (?) are optional.
type Account struct {
	XPub           *string `json:"xpub"`
	Index          *int    `json:"index"`
	DerivationMode *string `json:"derivationMode"` // standard, segwit, or native_segwit
	DerivationPath *string `json:"derivationPath"` // (?) Will override libcore defaults
	Birthday       *date   `json:"birthday"`       // (?) Earliest known creation date (YYYY/MM/DD)
}

// Configuration is a struct to model the JSON configuration
// of the project, stored in ~/.sats.json file.
type Configuration struct {
	Accounts    []Account `json:"accounts"`
	RPCURL      *string   `json:"rpcURL"`
	RPCUser     *string   `json:"rpcUser"`
	RPCPassword *string   `json:"rpcPassword"`
	RPCTLS      bool      `json:"rpcTLS"`
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

func (c *Configuration) Validate() {
	validateStringField("rpcURL", c.RPCURL)
	validateStringField("rpcUser", c.RPCUser)
	validateStringField("rpcPassword", c.RPCPassword)

	if c.Accounts == nil {
		log.WithFields(log.Fields{
			"accounts": c.Accounts,
		}).Fatal("Config validation failed")
	}

	for _, account := range c.Accounts {
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
