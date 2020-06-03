package transport

import (
	"ledger-sats-stack/pkg/types"

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
					if contains(addresses, inputAddress) && !contains(result, tx.TxID) {
						result = append(result, tx.TxID)
						break
					}
				}
			}

			if contains(addresses, tx.Address) && !contains(result, tx.TxID) {
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

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
