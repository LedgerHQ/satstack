package httpd

import (
	"ledger-sats-stack/pkg/handlers"
	"ledger-sats-stack/pkg/transport"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

func GetRouter(xrpc transport.XRPC, db *bolt.DB) *gin.Engine {
	engine := gin.Default()

	baseRouter := engine.Group("blockchain/v3")
	{
		baseRouter.GET("explorer/_health", handlers.GetHealth(xrpc))
		baseRouter.GET("syncToken", handlers.GetSyncToken(xrpc, db))
		baseRouter.DELETE("syncToken", handlers.DeleteSyncToken(xrpc, db))
	}

	var currency string
	info, _ := xrpc.GetBlockChainInfo()
	switch info.Chain {
	case "regtest", "test":
		currency = "btc_testnet"
	case "main":
		currency = "btc"
	}
	currencyRouter := baseRouter.Group(currency)
	{
		currencyRouter.GET("fees", handlers.GetFees(xrpc))
	}

	blocksRouter := currencyRouter.Group("/blocks")
	{
		blocksRouter.GET(":block", handlers.GetBlock(xrpc))
	}

	transactionsRouter := currencyRouter.Group("/transactions")
	{
		transactionsRouter.GET(":hash", handlers.GetTransaction(xrpc))
		transactionsRouter.GET(":hash/hex", handlers.GetTransactionHex(xrpc))
		transactionsRouter.POST("send", handlers.SendTransaction(xrpc))
	}

	addressesRouter := currencyRouter.Group("/addresses")
	{
		addressesRouter.GET(":addresses/transactions", handlers.GetAddresses(xrpc))
	}

	return engine
}

// GetXRPC initializes an XRPC stuct that embeds a btcd RPC client.
func GetXRPC(host string, user string, pass string, tls bool) transport.XRPC {
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

	return transport.XRPC{Client: client}
}
