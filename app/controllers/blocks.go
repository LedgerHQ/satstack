package controllers

import (
	"net/http"

	blocksRPC "ledger-sats-stack/app/rpc/blocks"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gin-gonic/gin"
)

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

		block, err := blocksRPC.GetBlock(blockRef, client)
		if err != nil {
			ctx.JSON(http.StatusNotFound, err)
			return
		}

		switch blockRef {
		case "current":
			ctx.JSON(http.StatusOK, block)
		default:
			ctx.JSON(http.StatusOK, []*blocksRPC.BlockContainer{block})
		}
	}
}
