package transport

import (
	"fmt"

	"ledger-sats-stack/pkg/types"
	"ledger-sats-stack/pkg/utils"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
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
func (w Wire) GetBlock(blockRef string) (*BlockContainer, error) {
	rawBlockHash, err := w.getBlockHashByReference(blockRef)
	if err != nil {
		return nil, err
	}

	block, err := w.getBlockByHash(rawBlockHash)
	if err != nil {
		return nil, err
	}

	return block, nil
}

func (w Wire) GetBlockHeightByHash(hash string) int64 {
	hashRaw, err := chainhash.NewHashFromStr(hash)
	if err != nil {
		// TODO: Log an error here
		return -1
	}

	rawBlock, err := w.GetBlockVerbose(hashRaw)

	if err != nil {
		// TODO: Log an error here
		return -1

	}
	return rawBlock.Height
}
