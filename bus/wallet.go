package bus

import (
	"fmt"

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

func ImportDescriptors(client *rpcclient.Client, descriptors []descriptor) error {
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

	opts := &btcjson.ImportMultiOptions{Rescan: true}

	results, err := client.ImportMulti(requests, opts)
	if err != nil {
		return err
	}

	var hasError bool

	for idx, result := range results {
		fields := log.WithFields(log.Fields{
			"descriptor": *requests[idx].Descriptor,
		})

		if result.Error != nil {
			fields.Error("Failed to import descriptor")
			hasError = true
		}

		if result.Success {
			fields.Debug("Import descriptor successfully")
		}
	}

	if hasError {
		return fmt.Errorf("importmulti RPC failed")
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

	var tx *types.Transaction

	switch b.TxIndex {
	case true:
		txRaw, err := b.mainClient.GetRawTransaction(chainHash)
		if err != nil {
			return nil, err
		}

		tx = protocol.DecodeMsgTx(txRaw.MsgTx(), b.Params)

	case false:
		txRaw, err := b.mainClient.GetTransactionWatchOnly(chainHash, true)
		if err != nil {
			return nil, err
		}

		tx, err = protocol.DecodeRawTransaction(txRaw.Hex, b.Params)
		if err != nil {
			return nil, err
		}
	}

	if b.Cache != nil {
		b.Cache.Set(hash, tx, cache.NoExpiration)
	}

	return tx, nil
}
