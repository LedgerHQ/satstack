package bus

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/chaincfg/chainhash"

	"github.com/btcsuite/btcd/btcjson"
	log "github.com/sirupsen/logrus"
)

const listTransactionsBatchSize = 1000

func (b *Bus) ListTransactions() []btcjson.ListTransactionsResult {
	var result []btcjson.ListTransactionsResult
	offset := 0

	for {
		txs, err := b.client.ListTransactionsCountFromWatchOnly("*", listTransactionsBatchSize, offset)
		if err != nil {
			log.WithFields(log.Fields{
				"error":     err,
				"batchSize": listTransactionsBatchSize,
				"offset":    offset,
			}).Error("Failed to list transactions")

			// return whatever we have so far (possibly empty slice)
			return result
		}

		if len(txs) == 0 {
			// no more transactions
			return result
		}

		result = append(result, txs...)
		offset += listTransactionsBatchSize
	}
}

func (b *Bus) GetTransaction(hash *chainhash.Hash) (*btcjson.TxRawResult, error) {
	switch b.TxIndex {
	case true:
		txRaw, err := b.client.GetRawTransactionVerbose(hash)
		if err != nil {
			return nil, err
		}

		return txRaw, nil
	default:
		tx, err := b.client.GetTransactionWatchOnly(hash, true)
		if err != nil {
			return nil, err
		}

		serializedTx, err := hex.DecodeString(tx.Hex)
		if err != nil {
			return nil, err
		}

		txRaw, err := b.client.DecodeRawTransaction(serializedTx)
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
	return b.client.GetAddressInfo(address)
}
