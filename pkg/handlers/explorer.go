package handlers

import (
	"log"
	"net/http"

	"ledger-sats-stack/pkg/transport"

	"github.com/gin-gonic/gin"
)

func GetHealth(wire transport.Wire) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		err := wire.GetHealth()
		if err != nil {
			log.Fatal(err)
			ctx.JSON(http.StatusServiceUnavailable, err)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"Status": "OK"})
	}
}
