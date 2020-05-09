package main

import (
	"os"

	"ledger-sats-stack/pkg/httpd"

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

	wire := httpd.GetWire(
		os.Getenv("BITCOIND_RPC_HOST"),
		os.Getenv("BITCOIND_RPC_USER"),
		os.Getenv("BITCOIND_RPC_PASSWORD"),
		os.Getenv("BITCOIND_RPC_ENABLE_TLS") == "true",
	)
	defer wire.Shutdown()

	db, err := bolt.Open("sats.db", 0666, nil)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	engine := httpd.GetRouter(wire, db)
	engine.Run()
}
