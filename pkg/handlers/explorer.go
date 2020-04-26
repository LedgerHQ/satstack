package handlers

import (
	"log"
	"net/http"

	"ledger-sats-stack/pkg/transport"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gin-gonic/gin"
)

func GetHealth(client *rpcclient.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		wire := transport.Wire{client}

		err := wire.GetHealth()
		if err != nil {
			log.Fatal(err)
			ctx.JSON(http.StatusServiceUnavailable, err)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"Status": "OK"})
	}
}
