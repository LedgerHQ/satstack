package httpd

import (
	"ledger-sats-stack/pkg/handlers"
	"ledger-sats-stack/pkg/transport"

	"github.com/gin-gonic/gin"
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