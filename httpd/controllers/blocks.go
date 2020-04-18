package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gin-gonic/gin"
)

// Block is a struct representing blocks
type Block struct {
	Hash         string   `json:"hash"`   // 0x prefixed
	Height       int64    `json:"height"` // integer
	Time         string   `json:"time"`   // RFC3339 format
	Transactions []string `json:"txs"`    // 0x prefixed
}

func (block *Block) init(hash string, height int64, timestamp int64, transactions []string) {
	block.Hash = fmt.Sprintf("0x%s", hash)
	block.Height = height
	block.Time = time.Unix(timestamp, 0).Format(time.RFC3339)

	for idx, transaction := range transactions {
		transactions[idx] = fmt.Sprintf("0x%s", transaction)
	}

	block.Transactions = transactions
}

// GetBlock gets the current block, or a block by height or hash.
// Examples:
//   - current    -> get number of blocks in longest blockchain
//   - 0xdeadbeef -> get block(s) by hash
//   - 626553     -> get block(s) by height
//
// Except for the case where the block reference is "current", the response is
// a list of 1 element.
func GetBlock(client *rpcclient.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		blockRef := ctx.Param("block")

		blockHashForeign, err := getBlockHashByReference(blockRef, client)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, err)
			return
		}

		blockForeign, err := client.GetBlockVerbose(blockHashForeign)
		if err != nil {
			ctx.JSON(http.StatusNotFound, err)
			return
		}

		block := new(Block)
		block.init(
			blockForeign.Hash,
			blockForeign.Height,
			blockForeign.Time,
			blockForeign.Tx,
		)

		ctx.JSON(http.StatusOK, block)

	}
}

func getBlockHashByReference(blockRef string, client *rpcclient.Client) (*chainhash.Hash, error) {
	switch {
	case blockRef == "current":
		return client.GetBestBlockHash()

	case strings.HasPrefix(blockRef, "0x"):
		return chainhash.NewHashFromStr(strings.TrimLeft(blockRef, "0x"))

	default:
		{
			blockHeight, err := strconv.ParseInt(blockRef, 10, 64)

			switch err {
			case nil:
				return client.GetBlockHash(blockHeight)

			default:
				return nil, fmt.Errorf("Invalid block '%s'", blockRef)
			}
		}

	}
}
