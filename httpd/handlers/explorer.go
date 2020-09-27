package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/onyb/sat-stack/httpd/svc"

	"github.com/gin-gonic/gin"
)

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

func GetTimestamp() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"timestamp": time.Now().Unix(),
		})
	}
}

func GetStatus(s svc.ExplorerService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		status, err := s.GetStatus()
		if err != nil {
			ctx.JSON(http.StatusServiceUnavailable, err)
			return
		}

		ctx.JSON(http.StatusOK, status)
	}
}
