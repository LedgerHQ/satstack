package handlers

import (
	"net/http"

	"ledger-sats-stack/pkg/transport"

	"github.com/gin-gonic/gin"
)

// GetTransaction is a gin handler (factory) to query transaction details
// by hash parameter.
func GetTransaction(xrpc transport.XRPC) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		txHash := ctx.Param("hash")

		transaction, err := xrpc.GetTransaction(txHash)
		if err != nil {
			ctx.JSON(http.StatusNotFound, err)
			return
		}

		ctx.JSON(http.StatusOK, []*transport.TransactionContainer{transaction})
	}
}

// GetTransactionHex is a gin handler (factory) to query transaction hex
// by hash parameter.
func GetTransactionHex(xrpc transport.XRPC) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		txHash := ctx.Param("hash")

		txHex, err := xrpc.GetTransactionHexByHash(txHash)
		if err != nil {
			ctx.JSON(http.StatusNotFound, err)
			return
		}

		response := gin.H{
			"transaction_hash": txHash,
			"hex":              txHex,
		}

		ctx.JSON(http.StatusOK, []gin.H{response})
	}
}
