package controllers

import (
	"net/http"

	"ledger-sats-stack/app/services"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gin-gonic/gin"
)

// GetTransaction is a gin handler (factory) to query transaction details
// by hash parameter.
func GetTransaction(client *rpcclient.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		txHash := ctx.Param("hash")
		transaction, err := services.GetTransaction(txHash, client)
		if err != nil {
			ctx.JSON(http.StatusNotFound, err)
			return
		}
		ctx.JSON(http.StatusOK, []*services.TransactionContainer{transaction})
	}
}

// GetTransactionHex is a gin handler (factory) to query transaction hex
// by hash parameter.
func GetTransactionHex(client *rpcclient.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		txHash := ctx.Param("hash")
		txHex, err := services.GetTransactionHexByHash(client, txHash)
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
