package bus

import (
	"errors"

	"github.com/ledgerhq/satstack/protocol"
	"github.com/ledgerhq/satstack/types"

	"github.com/ledgerhq/satstack/utils"

	"github.com/patrickmn/go-cache"

	"github.com/btcsuite/btcd/chaincfg/chainhash"

	"github.com/btcsuite/btcd/btcjson"
	log "github.com/sirupsen/logrus"
)

func (b *Bus) ListTransactions(blockHash *string) ([]btcjson.ListTransactionsResult, error) {
	client := b.getClient()
	defer b.recycleClient(client)

	var blockHashNative *chainhash.Hash
	if blockHash != nil {
		var err error
		blockHashNative, err = utils.ParseChainHash(*blockHash)
		if err != nil {
			return nil, err
		}
	}

	txs, err := client.ListSinceBlockMinConfWatchOnly(blockHashNative, 1, true)
	if err != nil {
		return nil, err
	}

	return txs.Transactions, nil
}

func (b *Bus) GetTransactionHex(hash *chainhash.Hash) (string, error) {
	client := b.getClient()
	defer b.recycleClient(client)

	tx, err := client.GetTransactionWatchOnly(hash, true)
	if err != nil {
		return "", err
	}

	return tx.Hex, nil
}

func (b *Bus) GetAddressInfo(address string) (*btcjson.GetAddressInfoResult, error) {
	client := b.getClient()
	defer b.recycleClient(client)

	return client.GetAddressInfo(address)
}

func (b *Bus) GetWalletInfo() (*btcjson.GetWalletInfoResult, error) {
	client := b.getClient()
	defer b.recycleClient(client)

	return client.GetWalletInfo()
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

	client := b.getClient()
	defer b.recycleClient(client)

	results, err := client.ImportMulti(requests, &btcjson.ImportMultiOptions{Rescan: true})
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

func (b *Bus) GetTransaction(hash *chainhash.Hash) (*types.Transaction, error) {
	if b.Cache != nil { // Cache has been enabled at the svc level
		if tx, found := b.Cache.Get(hash.String()); found {
			return tx.(*types.Transaction), nil
		}
	}

	client := b.getClient()
	defer b.recycleClient(client)

	txRaw, err := client.GetTransactionWatchOnly(hash, true)
	if err != nil {
		return nil, err
	}

	tx, err := protocol.DecodeRawTransaction(txRaw.Hex, b.Params)
	if err != nil {
		return nil, err
	}

	if b.Cache != nil {
		b.Cache.Set(hash.String(), tx, cache.NoExpiration)
	}

	return tx, nil
}
