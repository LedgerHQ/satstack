package handlers

import (
	"ledger-sats-stack/httpd/svc"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetAddresses(s svc.AddressesService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		param := ctx.Param("addresses")
		addressList := strings.Split(param, ",")

		addresses, err := s.GetAddresses(addressList)
		if err != nil {
			ctx.JSON(http.StatusNotFound, err)
			return
		}

		ctx.JSON(http.StatusOK, addresses)
	}
}
