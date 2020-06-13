package httpd

import (
	"ledger-sats-stack/httpd/handlers"
	"ledger-sats-stack/httpd/svc"

	"github.com/gin-gonic/gin"
)

func GetRouter(s *svc.Service) *gin.Engine {
	engine := gin.Default()

	baseRouter := engine.Group("blockchain/v3")
	{
		baseRouter.GET("explorer/_health", handlers.GetHealth(s))
	}

	currencyRouter := baseRouter.Group(s.Bus.Currency)
	{
		currencyRouter.GET("fees", handlers.GetFees(s))
	}

	blocksRouter := currencyRouter.Group("/blocks")
	{
		blocksRouter.GET(":block", handlers.GetBlock(s))
	}

	transactionsRouter := currencyRouter.Group("/transactions")
	{
		transactionsRouter.GET(":hash", handlers.GetTransaction(s))
		transactionsRouter.GET(":hash/hex", handlers.GetTransactionHex(s))
		transactionsRouter.POST("send", handlers.SendTransaction(s))
	}

	addressesRouter := currencyRouter.Group("/addresses")
	{
		addressesRouter.GET(":addresses/transactions", handlers.GetAddresses(s))
	}

	return engine
}
