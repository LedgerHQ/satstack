package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ledgerhq/satstack/httpd/svc"
	log "github.com/sirupsen/logrus"
)

// GetTransactionHex is a gin handler (factory) to query transaction hex
// by hash parameter.
func GetTransactionHex(s svc.TransactionsService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		txHash := ctx.Param("hash")

		txHex, err := s.GetTransactionHex(txHash)
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

func SendTransaction(s svc.TransactionsService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var request struct {
			Transaction string `json:"tx" binding:"required"`
		}

		if err := ctx.BindJSON(&request); err != nil {
			log.Error("Failed to bind JSON request")
			ctx.JSON(http.StatusBadRequest, err)
			return
		}

		txHash, err := s.SendTransaction(request.Transaction)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, err)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"result": txHash,
		})
	}
}
