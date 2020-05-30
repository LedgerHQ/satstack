package httpd

import (
	bolt "go.etcd.io/bbolt"
	"ledger-sats-stack/pkg/handlers"
	"ledger-sats-stack/pkg/transport"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func GetRouter(wire transport.Wire, db *bolt.DB) *gin.Engine {
	engine := gin.Default()

	baseRouter := engine.Group("blockchain/v3")
	baseRouter.GET("explorer/_health", handlers.GetHealth(wire))
	baseRouter.GET("syncToken", handlers.GetSyncToken(wire, db))
	baseRouter.DELETE("syncToken", handlers.DeleteSyncToken(wire, db))

	var currency string
	info, _ := wire.GetBlockChainInfo()
	switch info.Chain {
	case "regtest", "test":
		currency = "btc_testnet"
	case "main":
		currency = "btc"
	}
	currencyRouter := baseRouter.Group(currency)
	currencyRouter.GET("fees", handlers.GetFees(wire))

	blocksRouter := currencyRouter.Group("/blocks")
	blocksRouter.GET(":block", handlers.GetBlock(wire))

	transactionsRouter := currencyRouter.Group("/transactions")
	transactionsRouter.GET(":hash", handlers.GetTransaction(wire))
	transactionsRouter.GET(":hash/hex", handlers.GetTransactionHex(wire))

	return engine
}

// GetWire initializes a Wire stuct that embeds an RPC client.
func GetWire(host string, user string, pass string, tls bool) transport.Wire {
	connCfg := &rpcclient.ConnConfig{
		Host:         host,
		User:         user,
		Pass:         pass,
		HTTPPostMode: true,
		DisableTLS:   !tls,
	}
	// The notification parameter is nil since notifications are not
	// supported in HTTP POST mode.
	client, err := rpcclient.New(connCfg, nil)
	if client == nil || err != nil {
		log.WithFields(log.Fields{
			"host": host,
			"user": user,
			"TLS":  tls,
		}).Fatal("Failed to initialize RPC client.")
	}

	info, err := client.GetBlockChainInfo()

	if info == nil || err != nil {
		log.WithFields(log.Fields{
			"host": host,
			"user": user,
			"TLS":  tls,
		}).Fatal("Failed to connect to RPC server.")
	}

	log.WithFields(log.Fields{
		"chain":         info.Chain,
		"blocks":        info.Blocks,
		"bestblockhash": info.BestBlockHash,
	}).Info("RPC connection established.")

	return transport.Wire{Client: client}
}
