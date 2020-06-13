package bus

import (
	"fmt"
	"ledger-sats-stack/pkg/types"
	"ledger-sats-stack/pkg/utils"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

func (b *Bus) GetBestBlockHash() (*chainhash.Hash, error) {
	return b.Client.GetBestBlockHash()
}

func (b *Bus) GetBlockHash(height int64) (*chainhash.Hash, error) {
	return b.Client.GetBlockHash(height)
}

func (b *Bus) GetBlock(hash *chainhash.Hash) (*types.Block, error) {
	nativeBlock, err := b.Client.GetBlockVerbose(hash)
	if err != nil {
		return nil, err
	}

	transactions := make([]string, len(nativeBlock.Tx))
	for idx, transaction := range nativeBlock.Tx {
		transactions[idx] = fmt.Sprintf("0x%s", transaction)
	}

	block := types.Block{
		Hash:         fmt.Sprintf("0x%s", nativeBlock.Hash),
		Height:       nativeBlock.Height,
		Time:         utils.ParseUnixTimestamp(nativeBlock.Time),
		Transactions: &transactions,
	}

	return &block, nil
}

func (b *Bus) GetBlockChainInfo() (*btcjson.GetBlockChainInfoResult, error) {
	return b.Client.GetBlockChainInfo()
}
