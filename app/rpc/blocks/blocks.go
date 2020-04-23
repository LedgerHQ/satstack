package blocks

import (
	"fmt"
	"strconv"
	"strings"

	"ledger-sats-stack/app/types"
	"ledger-sats-stack/app/utils"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
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
func GetBlock(blockRef string, client *rpcclient.Client) (*BlockContainer, error) {
	rawBlockHash, err := getBlockHashByReference(blockRef, client)
	if err != nil {
		return nil, err
	}

	block, err := getBlockByHash(rawBlockHash, client)
	if err != nil {
		return nil, err
	}

	return block, nil
}

func getBlockByHash(hash *chainhash.Hash, client *rpcclient.Client) (*BlockContainer, error) {
	rawBlock, err := client.GetBlockVerbose(hash)
	if err != nil {
		return nil, err
	}

	block := new(BlockContainer)
	block.init(rawBlock)
	return block, nil
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

func GetBlockHeightByHash(client *rpcclient.Client, hash string) int64 {
	hashRaw, err := chainhash.NewHashFromStr(hash)
	if err != nil {
		// TODO: Log an error here
		return -1
	}

	rawBlock, err := client.GetBlockVerbose(hashRaw)

	if err != nil {
		// TODO: Log an error here
		return -1

	}
	return rawBlock.Height
}
