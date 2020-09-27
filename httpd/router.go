package httpd

import (
	"github.com/onyb/sat-stack/httpd/handlers"
	"github.com/onyb/sat-stack/httpd/svc"

	"github.com/gin-gonic/gin"
)

func GetRouter(s *svc.Service) *gin.Engine {
	engine := gin.Default()

	engine.GET("timestamp", handlers.GetTimestamp())

	// We support both Ledger Blockchain Explorer v2 and v3. The version here
	// is irrelevant.
	baseRouter := engine.Group("blockchain/:version")
	{
		baseRouter.GET("explorer/_health", handlers.GetHealth(s))
		baseRouter.GET("explorer/status", handlers.GetStatus(s))
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
