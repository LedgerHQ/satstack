package transport

import (
	"ledger-sats-stack/pkg/types"
	"ledger-sats-stack/pkg/utils"

	log "github.com/sirupsen/logrus"
)

const listTransactionsBatchSize = 1000

func (x XRPC) GetAddresses(addresses []string) (types.Addresses, error) {
	txIDs := x.getWalletTransactions(addresses)

	var txs []types.Transaction
	for _, txID := range txIDs {
		tx, err := x.GetTransaction(txID)
		if err != nil {
			log.Error(err)
			continue
		}
		txs = append(txs, tx.Transaction)
	}

	return types.Addresses{
		Truncated:    false,
		Transactions: txs,
	}, nil
}

func (x XRPC) getWalletTransactions(addresses []string) []string {
	var result []string

	offset := 0
	for {
		partialTxs, err := x.ListTransactionsCountFromWatchOnly("*", listTransactionsBatchSize, offset)
		if err != nil {
			log.Error(err)
		}

		if len(partialTxs) == 0 {
			// no more transactions
			break
		}

		for _, tx := range partialTxs {
			if tx.Category == "send" {
				tx2, err := x.GetTransaction(tx.TxID)
				if err != nil {
					log.Error(err)
				}

				for _, inputAddress := range getTransactionInputAddresses(tx2.Transaction) {
					if utils.Contains(addresses, inputAddress) && !utils.Contains(result, tx.TxID) {
						result = append(result, tx.TxID)
						break
					}
				}
			}

			if utils.Contains(addresses, tx.Address) && !utils.Contains(result, tx.TxID) {
				result = append(result, tx.TxID)
			}
		}

		offset += listTransactionsBatchSize
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
