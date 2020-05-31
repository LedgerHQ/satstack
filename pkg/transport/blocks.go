package transport

import (
	"fmt"

	"ledger-sats-stack/pkg/types"
	"ledger-sats-stack/pkg/utils"

	"github.com/btcsuite/btcd/btcjson"
	log "github.com/sirupsen/logrus"
)

// BlockContainer is a wrapper type to define an init method for
// BlockWithTransactions
type BlockContainer struct {
	types.BlockWithTransactions
}

func (block *BlockContainer) init(rawBlock *btcjson.GetBlockVerboseResult) {
	block.Hash = fmt.Sprintf("0x%s", rawBlock.Hash)
	block.Height = rawBlock.Height
	block.Time = utils.ParseUnixTimestamp(rawBlock.Time)
	transactions := make([]string, len(rawBlock.Tx))
	for idx, transaction := range rawBlock.Tx {
		transactions[idx] = fmt.Sprintf("0x%s", transaction)
	}
	block.Transactions = transactions
}

// GetBlock is a service method to get a Block by blockRef
func (x XRPC) GetBlock(blockRef string) (*BlockContainer, error) {
	rawBlockHash, err := x.getBlockHashByReference(blockRef)
	if err != nil {
		return nil, err
	}

	block, err := x.getBlockByHash(rawBlockHash)
	if err != nil {
		return nil, err
	}

	return block, nil
}

func (x XRPC) GetBlockHeightByHash(hash string) int64 {
	chainHash, err := utils.ParseChainHash(hash)
	if err != nil {
		log.WithFields(log.Fields{
			"hash":  hash,
			"error": err,
		}).Errorf("Failed to parse block hash")
		return -1
	}

	rawBlock, err := x.GetBlockVerbose(chainHash)

	if err != nil {
		log.WithFields(log.Fields{
			"hash":  hash,
			"error": err,
		}).Errorf("Failed to get block")
		return -1
	}
	return rawBlock.Height
}
