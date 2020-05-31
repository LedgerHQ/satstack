package handlers

import (
	"net/http"
	"strings"

	"ledger-sats-stack/pkg/transport"

	"github.com/gin-gonic/gin"
)

func GetAddresses(xrpc transport.XRPC) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		param := ctx.Param("addresses")
		addressList := strings.Split(param, ",")

		addresses, err := xrpc.GetAddresses(addressList)
		if err != nil {
			ctx.JSON(http.StatusNotFound, err)
			return
		}

		ctx.JSON(http.StatusOK, addresses)
	}
}
