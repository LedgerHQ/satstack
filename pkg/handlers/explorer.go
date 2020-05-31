package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"ledger-sats-stack/pkg/transport"

	"github.com/gin-gonic/gin"
	uuid2 "github.com/google/uuid"
	bolt "go.etcd.io/bbolt"
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

func GetSyncToken(xrpc transport.XRPC, db *bolt.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		uuid, err := uuid2.NewRandom()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, err)
			return
		}

		err = db.Update(func(tx *bolt.Tx) error {
			bucket, err := tx.CreateBucketIfNotExists([]byte("SyncTokens"))
			if err != nil {
				return err
			}

			err = bucket.Put([]byte(uuid.String()), []byte("*"))
			return err
		})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, err)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"token": uuid.String()})
	}
}

func DeleteSyncToken(xrpc transport.XRPC, db *bolt.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.GetHeader("X-LedgerWallet-SyncToken")

		if token == "" {
			ctx.String(http.StatusBadRequest, "No X-LedgerWallet-SyncToken passed in headers")
			return
		}

		err := db.Update(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte("SyncTokens"))
			tokenBytes := bucket.Get([]byte(token))
			if tokenBytes == nil {
				return errors.New(fmt.Sprintf("Token %s was not found", token))
			}
			err := bucket.Delete([]byte(token))
			return err
		})
		if err != nil {
			ctx.String(http.StatusNotFound, err.Error())
			return
		}

		ctx.String(http.StatusOK, fmt.Sprintf("Token %s was deleted", token))
	}
}
