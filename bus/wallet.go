package bus

import (
	"encoding/hex"
	"errors"

	"github.com/btcsuite/btcd/chaincfg/chainhash"

	"github.com/btcsuite/btcd/btcjson"
	log "github.com/sirupsen/logrus"
)

func (b *Bus) ListTransactions(blockHash *chainhash.Hash) ([]btcjson.ListTransactionsResult, error) {
	txs, err := b.Client.ListSinceBlockMinConfWatchOnly(nil, 1, true)
	if err != nil {
		return nil, err
	}

	return txs.Transactions, nil
}

func (b *Bus) GetTransaction(hash *chainhash.Hash) (*btcjson.TxRawResult, error) {
	switch b.TxIndex {
	case true:
		txRaw, err := b.Client.GetRawTransactionVerbose(hash)
		if err != nil {
			return nil, err
		}

		return txRaw, nil
	default:
		tx, err := b.Client.GetTransactionWatchOnly(hash, true)
		if err != nil {
			return nil, err
		}

		serializedTx, err := hex.DecodeString(tx.Hex)
		if err != nil {
			return nil, err
		}

		txRaw, err := b.Client.DecodeRawTransaction(serializedTx)
		if err != nil {
			return nil, err
		}

		// The decoded transaction hex doesn't contain confirmation number and
		// block height/hash; it must be fetched from the GetTransactionResult
		// instance.
		txRaw.Confirmations = uint64(tx.Confirmations)
		txRaw.BlockHash = tx.BlockHash
		txRaw.Time = tx.Time
		txRaw.Blocktime = tx.BlockTime

		return txRaw, nil
	}
}

func (b *Bus) GetAddressInfo(address string) (*btcjson.GetAddressInfoResult, error) {
	return b.Client.GetAddressInfo(address)
}

func (b *Bus) ImportDescriptors(descriptors []string, depth int) error {
	var requests []btcjson.ImportMultiRequest
	for _, descriptor := range descriptors {
		requests = append(requests, btcjson.ImportMultiRequest{
			Descriptor: btcjson.String(descriptor),
			Range:      &btcjson.DescriptorRange{Value: []int{0, depth}},
			Timestamp:  btcjson.Timestamp{Value: 0}, // TODO: Use birthday here
			WatchOnly:  btcjson.Bool(true),
			KeyPool:    btcjson.Bool(false),
			Internal:   btcjson.Bool(false),
		})
	}

	log.WithFields(log.Fields{
		"rescan": true,
		"N":      len(requests),
	}).Info("Importing descriptors")

	results, err := b.Client.ImportMulti(requests, &btcjson.ImportMultiOptions{Rescan: true})
	if err != nil {
		return err
	}

	hasErrors := false

	for idx, result := range results {
		if result.Error != nil {
			log.WithFields(log.Fields{
				"descriptor": *requests[idx].Descriptor,
				"range":      requests[idx].Range.Value,
				"error":      result.Error,
			}).Error("Failed to import descriptor")
			hasErrors = true
		}

		if result.Warnings != nil {
			log.WithFields(log.Fields{
				"descriptor": *requests[idx].Descriptor,
				"range":      requests[idx].Range.Value,
				"warnings":   result.Warnings,
			}).Warn("Import output descriptor")
		}

		if result.Success {
			log.WithFields(log.Fields{
				"descriptor": *requests[idx].Descriptor,
				"range":      requests[idx].Range.Value,
			}).Info("Import descriptor successful")
		}
	}

	if hasErrors {
		return errors.New("importmulti RPC command failed")
	}

	return nil
}
