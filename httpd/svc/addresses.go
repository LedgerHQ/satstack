package svc

import (
	"github.com/ledgerhq/satstack/types"
	"github.com/ledgerhq/satstack/utils"

	"github.com/btcsuite/btcd/btcjson"

	log "github.com/sirupsen/logrus"
)

func (s *Service) GetAddresses(addresses []string, blockHash *string, blockHeight *int32) (types.Addresses, error) {
	// Cache the results of GetTransaction calls against the TxID. The avoids
	// wasteful querying of the Bitcoin node for the same TxID, within the
	// lifecycle of this function invocation.
	s.Bus.NewCache()
	defer s.Bus.FlushCache()

	blockchainInfo, err := s.Bus.GetBlockChainInfo()
	if err != nil {
		return types.Addresses{}, err
	}

	txResults, err := s.Bus.ListTransactions(blockHash)
	if err != nil {
		log.WithFields(log.Fields{
			"error":     err,
			"blockHash": nil,
		}).Error("Unable to fetch transaction")
		return types.Addresses{}, err
	}

	walletTxs := s.filterTransactionsByAddresses(addresses, txResults, blockchainInfo.Headers)

	txs := make([]types.Transaction, 0, len(walletTxs))
	for _, txn := range walletTxs {
		if blockHeight != nil {
			if !(txn.BlockHeight != nil || *txn.BlockHeight == *blockHeight) {
				continue
			}
		}

		block := blockFromTxResult(txn)
		tx, err := s.GetTransaction(txn.TxID, block, blockchainInfo.Headers)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"hash":  txn.TxID,
			}).Error("Unable to fetch transaction")

			s.Bus.Cache.Delete(txn.TxID)
			continue
		}

		// Be defensive here with the retrieved transaction, to avoid
		// nil pointer dereference.
		if tx != nil {
			txs = append(txs, *tx)
		}
	}

	return types.Addresses{
		Truncated:    false,
		Transactions: txs,
	}, nil
}

func (s *Service) filterTransactionsByAddresses(
	addresses []string, txs []btcjson.ListTransactionsResult, bestBlockHeight int32,
) []btcjson.ListTransactionsResult {
	var result []btcjson.ListTransactionsResult
	var visited []string

	for _, tx := range txs {
		if tx.Category == "send" {
			block := blockFromTxResult(tx)
			tx2, err := s.GetTransaction(tx.TxID, block, bestBlockHeight)
			if err != nil {
				log.WithFields(log.Fields{
					"error":    err,
					"hash":     tx.TxID,
					"category": tx.Category,
				}).Error("Failed to get wallet transaction")

				// abandon processing the current transaction
				continue
			}

			for _, inputAddress := range getTransactionInputAddresses(*tx2) {
				if utils.Contains(addresses, inputAddress) && !utils.Contains(visited, tx.TxID) {
					result = append(result, tx)
					visited = append(visited, tx.TxID)
					break
				}
			}
		}

		if utils.Contains(addresses, tx.Address) && !utils.Contains(visited, tx.TxID) {
			result = append(result, tx)
			visited = append(visited, tx.TxID)
		}
	}

	return result
}

func getTransactionInputAddresses(tx types.Transaction) []string {
	var result []string

	for _, txInput := range tx.Inputs {
		result = append(result, txInput.Address)
	}

	return result
}

func blockFromTxResult(tx btcjson.ListTransactionsResult) *types.Block {
	var height int64
	if tx.BlockHeight != nil {
		height = int64(*tx.BlockHeight)
	} else {
		height = -1
	}

	return &types.Block{
		Hash:   tx.BlockHash,
		Height: height,
		Time:   utils.ParseUnixTimestamp(tx.BlockTime),
	}
}
