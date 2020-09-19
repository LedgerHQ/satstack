package bus

import (
	"fmt"
	"github.com/btcsuite/btcd/rpcclient"
	"ledger-sats-stack/utils"
)

// Currency represents the currency type (btc) and the network params
// (Mainnet, testnet3, regtest, etc) in libcore parlance.
type Currency = string

const (
	Testnet Currency = "btc_testnet"
	Mainnet Currency = "btc"
)

// currencyFromChain is an adapter function to convert a chain (network) value
// to a Currency type that's understood by libcore.
func CurrencyFromChain(chain string) (Currency, error) {
	switch chain {
	case "regtest", "test":
		return Testnet, nil
	case "main":
		return Mainnet, nil
	default:
		return "", ErrUnrecognizedChain
	}
}

// TxIndexEnabled can be used to detect if the bitcoind server being connected
// has a transaction index (enabled by option txindex=1).
//
// It works by fetching the first (coinbase) transaction from the block at
// height 1. The genesis coinbase is normally not part of the transaction
// index, therefore we use block at height 1 instead of 0.
//
// If an irrecoverable error is encountered, it returns an error. In such
// cases, the caller may stop program execution.
func TxIndexEnabled(client *rpcclient.Client) (bool, error) {
	blockHash, err := client.GetBlockHash(1)
	if err != nil {
		return false, ErrFailedToGetBlock
	}

	block, err := client.GetBlockVerbose(blockHash)
	if err != nil {
		return false, ErrFailedToGetBlock
	}

	if len(block.Tx) == 0 {
		return false, fmt.Errorf("no transactions in block at height 1")
	}

	// Coinbase transaction in block at height 1
	tx, err := utils.ParseChainHash(block.Tx[0])
	if err != nil {
		return false, fmt.Errorf(
			"%s (%s): %w", ErrMalformedChainHash, block.Tx[0], err)
	}

	if _, err := client.GetRawTransaction(tx); err != nil {
		return false, nil
	}

	return true, nil
}
