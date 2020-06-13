package bus

import (
	"ledger-sats-stack/types"

	"github.com/btcsuite/btcutil"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

type Wallet interface {
	ListTransactions(blockHash *chainhash.Hash) ([]btcjson.ListTransactionsResult, error)
	GetTransaction(hash *chainhash.Hash) (*btcjson.TxRawResult, error)
	GetAddressInfo(address string) (*btcjson.GetAddressInfoResult, error)
	ImportDescriptors(descriptors []string, depth int) error
}

type Util interface {
	GetCanonicalDescriptor(descriptor string) (*string, error)
	DeriveAddress(descriptor string, index int) (*string, error)
	EstimateSmartFee(target int64, mode string) btcutil.Amount
}

type Chain interface {
	GetBlockChainInfo() (*btcjson.GetBlockChainInfoResult, error)
	GetBestBlockHash() (*chainhash.Hash, error)
	GetBlockHash(height int64) (*chainhash.Hash, error)
	GetBlock(hash *chainhash.Hash) (*types.Block, error)
}

type RawTransactions interface {
	SendTransaction(tx string) (*chainhash.Hash, error)
}

type Bridge interface {
	Chain
	Util
	RawTransactions
	Wallet
}
