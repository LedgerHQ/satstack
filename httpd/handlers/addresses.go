package handlers

import (
	"ledger-sats-stack/httpd/svc"
	"ledger-sats-stack/utils"
	"net/http"
	"sort"
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

		// FIXME: libcore relies on the order of the transactions, in order to
		//        correctly compute operation values (aka amounts). This order
		//        appears to be based on the ReceivedAt field, although it is
		//        not documented in the Ledger BE project.
		//
		//        The bug seems to manifest itself only on accounts with a
		//        large number of operations.
		sort.Slice(addresses.Transactions[:], func(i, j int) bool {
			iReceivedAt, iErr := utils.ParseRFC3339Timestamp(addresses.Transactions[i].ReceivedAt)
			jReceivedAt, jErr := utils.ParseRFC3339Timestamp(addresses.Transactions[j].ReceivedAt)

			if iErr != nil || jErr != nil {
				// Still a semi-reliable way of comparing RFC3339 timestamps.
				return addresses.Transactions[i].ReceivedAt < addresses.Transactions[j].ReceivedAt
			}

			return *iReceivedAt < *jReceivedAt
		})

		ctx.JSON(http.StatusOK, addresses)
	}
}
