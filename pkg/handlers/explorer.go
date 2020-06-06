package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"ledger-sats-stack/pkg/transport"

	"github.com/gin-gonic/gin"
)

func GetHealth(xrpc transport.XRPC) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		err := xrpc.GetHealth()
		if err != nil {
			ctx.JSON(http.StatusServiceUnavailable, err)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"Status": "OK"})
	}
}

func GetFees(xrpc transport.XRPC) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		blockCounts := ctx.QueryArray("block_count")
		mode := strings.ToUpper(ctx.Param("mode"))
		if mode == "" || (mode != "UNSET" && mode != "ECONOMICAL" && mode != "CONSERVATIVE") {
			mode = "CONSERVATIVE"
		}

		var blockCountsIntegers []int64
		for _, blockCount := range blockCounts {
			if value, err := strconv.ParseInt(blockCount, 10, 64); err == nil {
				blockCountsIntegers = append(blockCountsIntegers, value)
			}
		}

		if len(blockCountsIntegers) == 0 {
			blockCountsIntegers = append(blockCountsIntegers, 2, 3, 6)
		}

		fees := xrpc.GetSmartFeeEstimates(blockCountsIntegers, mode)
		ctx.JSON(http.StatusOK, fees)
	}
}
