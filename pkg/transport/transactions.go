package transport

import (
	log "github.com/sirupsen/logrus"
	. "ledger-sats-stack/pkg/types"
	"ledger-sats-stack/pkg/utils"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcutil"
)

// TransactionContainer is a wrapper type to define an init method for
// Transaction
type TransactionContainer struct {
	Transaction
}

func (txn *TransactionContainer) init(rawTx *btcjson.TxRawResult, utxos UTXOs, blockHeight int64) {
	txn.ID = rawTx.Txid
	txn.Hash = rawTx.Txid // !FIXME: Use rawTx.Hash, which can differ for witness transactions
	txn.ReceivedAt = utils.ParseUnixTimestamp(rawTx.Time)
	txn.LockTime = rawTx.LockTime

	vin := make([]Input, len(rawTx.Vin))
	sumVinValues := btcutil.Amount(0)
	vinHasCoinbase := false

	for idx, rawVin := range rawTx.Vin {
		inputIndex := idx

		if rawVin.IsCoinBase() {
			vin[idx] = Input{
				Coinbase:   rawVin.Coinbase,
				InputIndex: &inputIndex,
				Sequence:   rawVin.Sequence,
			}

			vinHasCoinbase = true
		} else {
			utxo := utxos[OutputIdentifier{Hash: rawVin.Txid, Index: rawVin.Vout}]
			outputIndex := rawVin.Vout

			vin[idx] = Input{
				OutputHash:  rawVin.Txid,
				OutputIndex: &outputIndex,
				InputIndex:  &inputIndex, // TODO: Find out if the order matters
				Value:       &utxo.Value,
				Address:     utxo.Address,
				ScriptSig:   &rawVin.ScriptSig.Hex,
				Sequence:    rawVin.Sequence,
			}
			if rawVin.HasWitness() {
				witness := rawVin.Witness
				vin[idx].Witness = &witness
			} else {
				vin[idx].Witness = &[]string{} // !FIXME: Coinbase txn can also have witness
			}

			sumVinValues += *vin[idx].Value
		}
	}
	txn.Inputs = vin

	vout := make([]Output, len(rawTx.Vout))
	sumVoutValues := btcutil.Amount(0)

	for idx, rawVout := range rawTx.Vout {
		outputValue := utils.ParseSatoshi(rawVout.Value) // !FIXME: Can panic
		outputIndex := rawVout.N
		vout[idx] = Output{
			OutputIndex: &outputIndex,
			Value:       &outputValue,
			ScriptHex:   rawVout.ScriptPubKey.Hex,
		}

		if len(rawVout.ScriptPubKey.Addresses) >= 1 {
			// ScriptPubKey can have multiple addresses for a multisig
			// transaction.
			log.WithFields(log.Fields{
				"addresses":   rawVout.ScriptPubKey.Addresses,
				"value":       outputValue,
				"outputIndex": rawVout.N,
			}).Warnf("Multisig transaction detected.")
			vout[idx].Address = rawVout.ScriptPubKey.Addresses[0]
		} else {
			// TODO: Document when this happens
		}

		sumVoutValues += *vout[idx].Value
	}
	txn.Outputs = vout

	txn.Block = Block{
		Hash:   rawTx.BlockHash,
		Height: blockHeight,
		Time:   utils.ParseUnixTimestamp(rawTx.Blocktime),
	}

	// ?XXX: Confirmations in Ledger Blockchain Explorer are always off by 1
	txn.Confirmations = rawTx.Confirmations - uint64(1)

	var fees btcutil.Amount

	if vinHasCoinbase {
		// Coinbase transaction have no fees
		fees = btcutil.Amount(0)
	} else {
		fees = sumVinValues - sumVoutValues
	}
	txn.Fees = &fees
}

// GetTransaction is a service function to query transaction details
// by hash parameter.
func (w Wire) GetTransaction(txHash string) (*TransactionContainer, error) {
	txRaw, err := w.getTransactionByHash(txHash)
	if err != nil {
		return nil, err
	}

	utxos, err := w.buildUTXOs(txRaw.Vin)
	if err != nil {
		return nil, err
	}

	blockHeight := w.GetBlockHeightByHash(txRaw.BlockHash)

	transaction := new(TransactionContainer)
	transaction.init(txRaw, utxos, blockHeight)
	return transaction, nil
}

// GetTransactionHexByHash is a service function to get hex encoded raw
// transaction by hash.
func (w Wire) GetTransactionHexByHash(txHash string) (string, error) {
	txRaw, err := w.getTransactionByHash(txHash)
	if err != nil {
		return "", err
	}
	return txRaw.Hex, nil
}
