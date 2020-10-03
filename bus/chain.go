package bus

import (
	"github.com/ledgerhq/satstack/types"
	"github.com/ledgerhq/satstack/utils"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

func (b *Bus) GetBestBlockHash() (*chainhash.Hash, error) {
	client := b.getClient()
	defer b.recycleClient(client)

	return client.GetBestBlockHash()
}

func (b *Bus) GetBlockHash(height int64) (*chainhash.Hash, error) {
	client := b.getClient()
	defer b.recycleClient(client)

	return client.GetBlockHash(height)
}

func (b *Bus) GetBlock(hash *chainhash.Hash) (*types.Block, error) {
	client := b.getClient()
	defer b.recycleClient(client)

	nativeBlock, err := client.GetBlockVerbose(hash)
	if err != nil {
		return nil, err
	}

	transactions := make([]string, len(nativeBlock.Tx))
	for idx, transaction := range nativeBlock.Tx {
		transactions[idx] = transaction
	}

	block := types.Block{
		Hash:         nativeBlock.Hash,
		Height:       nativeBlock.Height,
		Time:         utils.ParseUnixTimestamp(nativeBlock.Time),
		Transactions: &transactions,
	}

	return &block, nil
}

func (b *Bus) GetBlockChainInfo() (*btcjson.GetBlockChainInfoResult, error) {
	client := b.getClient()
	defer b.recycleClient(client)

	return client.GetBlockChainInfo()
}
