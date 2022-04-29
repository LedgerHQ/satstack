package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ledgerhq/satstack/httpd/svc"
)

func GetFees(s svc.ExplorerService) gin.HandlerFunc {
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

		fees := s.GetFees(blockCountsIntegers, mode)
		ctx.JSON(http.StatusOK, fees)
	}
}

func GetHealth(s svc.ExplorerService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		err := s.GetHealth()
		if err != nil {
			ctx.JSON(http.StatusServiceUnavailable, err)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"Status": "OK"})
	}
}

func GetNetwork(s svc.ExplorerService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, s.GetNetwork())
	}
}

func GetStatus(s svc.ExplorerService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, s.GetStatus())
	}
}

func GetTimestamp() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"timestamp": time.Now().Unix(),
		})
	}
}
