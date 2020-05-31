package handlers

import (
	"net/http"

	"ledger-sats-stack/pkg/transport"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
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

func SendTransaction(xrpc transport.XRPC) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var request struct {
			Transaction string `json:"tx" binding:"required"`
		}
		if err := ctx.BindJSON(&request); err != nil {
			log.Error("Failed to bind JSON request")
			ctx.JSON(http.StatusBadRequest, err)
			return
		}

		txHash, err := xrpc.SendTransaction(request.Transaction)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, err)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"result": *txHash,
		})
	}
}
