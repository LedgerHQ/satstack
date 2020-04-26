package handlers

import (
	"net/http"

	"ledger-sats-stack/pkg/transport"

	"github.com/gin-gonic/gin"
)

// GetTransaction is a gin handler (factory) to query transaction details
// by hash parameter.
func GetTransaction(wire transport.Wire) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		txHash := ctx.Param("hash")

		transaction, err := wire.GetTransaction(txHash)
		if err != nil {
			ctx.JSON(http.StatusNotFound, err)
			return
		}

		ctx.JSON(http.StatusOK, []*transport.TransactionContainer{transaction})
	}
}

// GetTransactionHex is a gin handler (factory) to query transaction hex
// by hash parameter.
func GetTransactionHex(wire transport.Wire) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		txHash := ctx.Param("hash")

		txHex, err := wire.GetTransactionHexByHash(txHash)
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
