package main

import (
	"encoding/json"
	"os"
	"path"

	"ledger-sats-stack/pkg/httpd"
	"ledger-sats-stack/pkg/types"

	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	bolt "go.etcd.io/bbolt"
)

func main() {
	log.SetFormatter(&prefixed.TextFormatter{
		TimestampFormat:  "2006/01/02 - 15:04:05",
		FullTimestamp:    true,
		QuoteEmptyFields: true,
		SpacePadding:     45,
	})

	xrpc := httpd.GetXRPC(
		os.Getenv("BITCOIND_RPC_HOST"),
		os.Getenv("BITCOIND_RPC_USER"),
		os.Getenv("BITCOIND_RPC_PASSWORD"),
		os.Getenv("BITCOIND_RPC_ENABLE_TLS") == "true",
	)
	defer xrpc.Shutdown()

	db, err := bolt.Open("sats.db", 0666, nil)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	accounts := loadAccountsConfig()
	_ = xrpc.ImportAccounts(accounts)

	engine := httpd.GetRouter(xrpc, db)
	engine.Run(":20000")
}

func loadAccountsConfig() []types.Account {
	home, err := homedir.Dir()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Cannot obtain user home directory")
	}

	configPath := path.Join(home, ".sats.json")

	file, err := os.Open(configPath)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Cannot open config file")
	}

	defer file.Close()

	decoder := json.NewDecoder(file)
	accounts := []types.Account{}

	err = decoder.Decode(&accounts)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Cannot decode accounts config JSON")
	}

	log.WithFields(log.Fields{
		"path":        configPath,
		"numAccounts": len(accounts),
	}).Info("Loaded config file")

	return accounts
}
