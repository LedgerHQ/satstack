package controllers

import (
	"fmt"
	"log"
	"net/http"
	"time"

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

// GetCurrentBlock gets the current number of blocks in the longest chain.
func GetCurrentBlock(client *rpcclient.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		blockHashForeign, err := client.GetBestBlockHash()
		if err != nil {
			log.Fatal(err)
		}

		blockForeign, err := client.GetBlockVerbose(blockHashForeign)
		if err != nil {
			log.Fatal(err)
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
