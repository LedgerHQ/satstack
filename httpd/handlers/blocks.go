package handlers

import (
	"net/http"

	"github.com/ledgerhq/satstack/httpd/svc"
	"github.com/ledgerhq/satstack/types"

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
func GetBlock(s svc.BlocksService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		blockRef := ctx.Param("block")

		block, err := s.GetBlock(blockRef)
		if err != nil {
			ctx.String(http.StatusNotFound, "text/plain", []byte(err.Error()))
			return
		}

		switch blockRef {
		case "current":
			ctx.JSON(http.StatusOK, block)
		default:
			ctx.JSON(http.StatusOK, []*types.Block{block})
		}
	}
}
