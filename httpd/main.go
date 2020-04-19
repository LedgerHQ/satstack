package main

import (
	"ledger-sats-stack/httpd/controllers"
	"log"
	"os"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gin-gonic/gin"
)

func main() {
	engine, client := Setup()
	defer client.Shutdown()

	engine.Run()
}

func Setup() (*gin.Engine, *rpcclient.Client) {
	engine := gin.Default()

	// Connect to local bitcoin core RPC server using HTTP POST mode.
	connCfg := &rpcclient.ConnConfig{
		Host:         os.Getenv("BITCOIND_RPC_HOST"),
		User:         os.Getenv("BITCOIND_RPC_USER"),
		Pass:         os.Getenv("BITCOIND_RPC_PASSWORD"),
		HTTPPostMode: true,
		DisableTLS:   os.Getenv("BITCOIND_RPC_ENABLE_TLS") != "true",
	}
	// The notification parameter is nil since notifications are not
	// supported in HTTP POST mode.
	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		log.Fatal(err)
	}

	engine.GET("/blockchain/v3/blocks/:block", controllers.GetBlock(client))
	engine.GET("/blockchain/v3/transactions/:hash", controllers.GetTransaction(client))

	return engine, client
}
