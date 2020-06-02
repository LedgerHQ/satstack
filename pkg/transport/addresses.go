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
			break
		}

		for _, tx := range partialTxs {
			if contains(addresses, tx.Address) && !contains(result, tx.TxID) {
				result = append(result, tx.TxID)
			}
		}

		offset += listTransactionsBatchSize
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
