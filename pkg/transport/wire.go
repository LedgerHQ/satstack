package transport

import (
	"fmt"
	"strconv"
	"strings"

	"ledger-sats-stack/pkg/types"
	"ledger-sats-stack/pkg/utils"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	log "github.com/sirupsen/logrus"
)

// Wire is a copper wire
type Wire struct {
	*rpcclient.Client
}

func (w Wire) getBlockByHash(hash *chainhash.Hash) (*BlockContainer, error) {
	rawBlock, err := w.GetBlockVerbose(hash)
	if err != nil {
		return nil, err
	}

	block := new(BlockContainer)
	block.init(rawBlock)
	return block, nil
}

func (w Wire) getBlockHashByReference(blockRef string) (*chainhash.Hash, error) {
	switch {
	case blockRef == "current":
		return w.GetBestBlockHash()

	case strings.HasPrefix(blockRef, "0x"), len(blockRef) == 64:
		// 256-bit hex string with or without 0x prefix
		return utils.ParseChainHash(blockRef)
	default:
		{
			// Either an int64 block height, or garbage input
			blockHeight, err := strconv.ParseInt(blockRef, 10, 64)

			switch err {
			case nil:
				return w.GetBlockHash(blockHeight)

			default:
				return nil, fmt.Errorf("Invalid block '%s'", blockRef)
			}
		}

	}
}

func (w Wire) buildUTXOs(vin []btcjson.Vin) (types.UTXOs, error) {
	utxos := make(types.UTXOs)
	utxoResults := make(map[types.OutputIdentifier]rpcclient.FutureGetRawTransactionVerboseResult)

	for _, inputRaw := range vin {
		if inputRaw.IsCoinBase() {
			continue
		}

		chainHash, err := utils.ParseChainHash(inputRaw.Txid)
		if err != nil {
			return nil, err
		}

		utxoResults[types.OutputIdentifier{Hash: inputRaw.Txid, Index: inputRaw.Vout}] = w.GetRawTransactionVerboseAsync(chainHash)
	}

	for utxoID, utxoResult := range utxoResults {
		tx, err := utxoResult.Receive()
		if err != nil {
			return nil, err
		}

		utxo, err := parseUTXO(tx, utxoID.Index)
		if err != nil {
			return nil, err
		}

		utxos[utxoID] = *utxo
	}

	return utxos, nil
}

func parseUTXO(tx *btcjson.TxRawResult, outputIndex uint32) (*types.UTXOData, error) {
	utxoRaw := tx.Vout[outputIndex]

	switch addresses := utxoRaw.ScriptPubKey.Addresses; len(addresses) {
	case 0:
		// TODO: Document when this happens
		return &types.UTXOData{
			Value:   utils.ParseSatoshi(utxoRaw.Value), // !FIXME: Can panic
			Address: "",                                // Will be omitted by the JSON serializer
		}, nil
	case 1:
		return &types.UTXOData{
			Value:   utils.ParseSatoshi(utxoRaw.Value),
			Address: addresses[0], // ?XXX: Investigate why we do this
		}, nil
	default:
		value := utils.ParseSatoshi(utxoRaw.Value) // !FIXME: Can panic
		log.WithFields(log.Fields{
			"addresses":   addresses,
			"value":       value,
			"outputIndex": outputIndex,
		}).Warn("Multisig transaction detected.")

		return &types.UTXOData{
			Value:   value,
			Address: addresses[0],
		}, nil
	}
}

// getTransactionByHash gets the transaction with the given hash.
// Supports transaction hashes with or without 0x prefix.
func (w Wire) getTransactionByHash(txHash string) (*btcjson.TxRawResult, error) {
	chainHash, err := utils.ParseChainHash(txHash)
	if err != nil {
		return nil, err
	}

	txRaw, err := w.GetRawTransactionVerbose(chainHash)

	if err != nil {
		return nil, err
	}
	return txRaw, nil
}
