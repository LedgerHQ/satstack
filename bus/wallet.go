package bus

import (
	"errors"
	"github.com/btcsuite/btcd/rpcclient"

	"github.com/ledgerhq/satstack/protocol"
	"github.com/ledgerhq/satstack/types"

	"github.com/ledgerhq/satstack/utils"

	"github.com/patrickmn/go-cache"

	"github.com/btcsuite/btcd/chaincfg/chainhash"

	"github.com/btcsuite/btcd/btcjson"
	log "github.com/sirupsen/logrus"
)

func (b *Bus) ListTransactions(blockHash *string) ([]btcjson.ListTransactionsResult, error) {
	var blockHashNative *chainhash.Hash
	if blockHash != nil {
		var err error
		blockHashNative, err = utils.ParseChainHash(*blockHash)
		if err != nil {
			return nil, err
		}
	}

	txs, err := b.mainClient.ListSinceBlockMinConfWatchOnly(blockHashNative, 1, true)
	if err != nil {
		return nil, err
	}

	return txs.Transactions, nil
}

func (b *Bus) GetTransactionHex(hash *chainhash.Hash) (string, error) {
	tx, err := b.mainClient.GetTransactionWatchOnly(hash, true)
	if err != nil {
		return "", err
	}

	return tx.Hex, nil
}

func (b *Bus) GetAddressInfo(address string) (*btcjson.GetAddressInfoResult, error) {
	return b.mainClient.GetAddressInfo(address)
}

func (b *Bus) GetWalletInfo() (*btcjson.GetWalletInfoResult, error) {
	return b.mainClient.GetWalletInfo()
}

func (b *Bus) ImportDescriptors(descriptors []descriptor) error {
	var requests []btcjson.ImportMultiRequest
	for _, descriptor := range descriptors {
		requests = append(requests, btcjson.ImportMultiRequest{
			Descriptor: btcjson.String(descriptor.Value),
			Range:      &btcjson.DescriptorRange{Value: []int{0, descriptor.Depth}},
			Timestamp:  btcjson.TimestampOrNow{Value: descriptor.Age},
			WatchOnly:  btcjson.Bool(true),
			KeyPool:    btcjson.Bool(false),
			Internal:   btcjson.Bool(false),
		})
	}

	log.WithFields(log.Fields{
		"rescan": true,
		"N":      len(requests),
	}).Info("Importing descriptors")

	results, err := b.mainClient.ImportMulti(requests, &btcjson.ImportMultiOptions{Rescan: true})
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

func (b *Bus) GetTransaction(hash string) (*types.Transaction, error) {
	if b.Cache != nil { // Cache has been enabled at the svc level
		if tx, found := b.Cache.Get(hash); found {
			return tx.(*types.Transaction), nil
		}
	}

	chainHash, err := utils.ParseChainHash(hash)
	if err != nil {
		return nil, err
	}

	txRaw, err := b.mainClient.GetTransactionWatchOnly(chainHash, true)
	if err != nil {
		return nil, err
	}

	tx, err := protocol.DecodeRawTransaction(txRaw.Hex, b.Params)
	if err != nil {
		return nil, err
	}

	if b.Cache != nil {
		b.Cache.Set(hash, tx, cache.NoExpiration)
	}

	return tx, nil
}

func (b *Bus) GetTransactionBatch(hashes []string) map[string]*types.Transaction {
	txnMap := make(map[string]*types.Transaction)

	if b.Cache != nil { // Cache has been enabled at the svc level
		for _, hash := range hashes {
			if tx, found := b.Cache.Get(hash); found {
				txnMap[hash] = tx.(*types.Transaction)
			}
		}
	}

	var txnFutures []rpcclient.FutureGetTransactionResult

	for _, hash := range hashes {
		if _, ok := txnMap[hash]; !ok {
			chainHash, err := utils.ParseChainHash(hash)
			if err != nil {
				log.WithFields(log.Fields{
					"err":  err,
					"hash": hash,
				}).Error("Failed to parse transaction hash")
				continue
			}

			txnFuture := b.batchClient.GetTransactionWatchOnlyAsync(
				chainHash, true)
			txnFutures = append(txnFutures, txnFuture)
		}
	}

	if err := b.batchClient.Send(); err != nil {
		log.WithField("err", err).Error("Failed to send batch RPC call")
		return txnMap
	}

	for _, txnFuture := range txnFutures {
		txRaw, err := txnFuture.Receive()
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("Failed to parse transaction future")
			continue
		}

		tx, err := protocol.DecodeRawTransaction(txRaw.Hex, b.Params)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
				"hex": txRaw.Hex,
			}).Error("Failed to parse raw transaction hex")
			continue
		}

		txnMap[tx.ID] = tx

		if b.Cache != nil {
			b.Cache.Set(tx.ID, tx, cache.NoExpiration)
		}
	}

	return txnMap
}
