package httpd

import (
	"ledger-sats-stack/pkg/handlers"
	"ledger-sats-stack/pkg/transport"

	"github.com/gin-gonic/gin"
)

func GetRouter(xrpc transport.XRPC) *gin.Engine {
	engine := gin.Default()

	baseRouter := engine.Group("blockchain/v3")
	{
		baseRouter.GET("explorer/_health", handlers.GetHealth(xrpc))
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
