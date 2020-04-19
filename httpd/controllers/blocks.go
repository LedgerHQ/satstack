package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcjson"
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

func (block *Block) init(rawBlock *btcjson.GetBlockVerboseResult) {
	block.Hash = fmt.Sprintf("0x%s", rawBlock.Hash)
	block.Height = rawBlock.Height
	block.Time = time.Unix(rawBlock.Time, 0).Format(time.RFC3339)

	transactions := make([]string, len(rawBlock.Tx))
	for idx, transaction := range rawBlock.Tx {
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

		rawBlockHash, err := getBlockHashByReference(blockRef, client)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, err)
			return
		}

		rawBlock, err := client.GetBlockVerbose(rawBlockHash)
		if err != nil {
			ctx.JSON(http.StatusNotFound, err)
			return
		}

		block := new(Block)
		block.init(rawBlock)

		switch blockRef {
		case "current":
			ctx.JSON(http.StatusOK, block)
		default:
			ctx.JSON(http.StatusOK, []*Block{block})
		}
	}
}

func getBlockHashByReference(blockRef string, client *rpcclient.Client) (*chainhash.Hash, error) {
	switch {
	case blockRef == "current":
		return client.GetBestBlockHash()

	case strings.HasPrefix(blockRef, "0x"):
		// 256-bit hex string with 0x prefix
		return chainhash.NewHashFromStr(strings.TrimLeft(blockRef, "0x"))
	case len(blockRef) == 64:
		// 256-bit hex string WITHOUT 0x prefix
		return chainhash.NewHashFromStr(blockRef)
	default:
		{
			// Either an int64 block height, or garbage input
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
