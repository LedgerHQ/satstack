package httpd

import (
	"github.com/ledgerhq/satstack/httpd/handlers"
	"github.com/ledgerhq/satstack/httpd/svc"

	"github.com/gin-gonic/gin"
)

func GetRouter(s *svc.Service) *gin.Engine {
	engine := gin.Default()

	engine.GET("timestamp", handlers.GetTimestamp())

	// controlRouter exposes endpoints that can be used to programmatically
	// control SatStack (for ex, from Ledger Live).
	controlRouter := engine.Group("control")
	{
		controlRouter.GET("descriptors/import", handlers.ImportAccounts(s))
		controlRouter.POST("descriptors/has", handlers.HasDescriptor(s))
	}

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
		transactionsRouter.GET(":hash/hex", handlers.GetTransactionHex(s))
		transactionsRouter.POST("send", handlers.SendTransaction(s))
	}

	addressesRouter := currencyRouter.Group("/addresses")
	{
		addressesRouter.GET(":addresses/transactions", handlers.GetAddresses(s))
	}

	return engine
}
