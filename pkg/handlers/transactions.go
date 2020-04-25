package handlers

import (
	"net/http"

	"ledger-sats-stack/pkg/transport"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gin-gonic/gin"
)

// GetTransaction is a gin handler (factory) to query transaction details
// by hash parameter.
func GetTransaction(client *rpcclient.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		txHash := ctx.Param("hash")
		wire := transport.Wire{client}

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
func GetTransactionHex(client *rpcclient.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		txHash := ctx.Param("hash")
		wire := transport.Wire{client}

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
