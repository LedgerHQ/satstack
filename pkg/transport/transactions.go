package transport

import (
	"bytes"
	"encoding/hex"
	"ledger-sats-stack/pkg/types"
	"ledger-sats-stack/pkg/utils"
	"time"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	log "github.com/sirupsen/logrus"
)

// TransactionContainer is a wrapper type to define an init method for
// Transaction
type TransactionContainer struct {
	types.Transaction
}

func (txn *TransactionContainer) init(rawTx *btcjson.TxRawResult, utxos types.UTXOs, block *types.Block) {
	txn.ID = rawTx.Txid
	txn.Hash = rawTx.Txid // !FIXME: Use rawTx.Hash, which can differ for witness transactions
	txn.LockTime = rawTx.LockTime

	vin := make([]types.Input, len(rawTx.Vin))
	sumVinValues := btcutil.Amount(0)
	vinHasCoinbase := false

	for idx, rawVin := range rawTx.Vin {
		inputIndex := idx

		if rawVin.IsCoinBase() {
			vin[idx] = types.Input{
				Coinbase:   rawVin.Coinbase,
				InputIndex: &inputIndex,
				Sequence:   rawVin.Sequence,
			}

			vinHasCoinbase = true
		} else {
			utxo := utxos[types.OutputIdentifier{Hash: rawVin.Txid, Index: rawVin.Vout}]
			outputIndex := rawVin.Vout

			vin[idx] = types.Input{
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

	vout := make([]types.Output, len(rawTx.Vout))
	sumVoutValues := btcutil.Amount(0)

	for idx, rawVout := range rawTx.Vout {
		outputValue := utils.ParseSatoshi(rawVout.Value) // !FIXME: Can panic
		outputIndex := rawVout.N
		vout[idx] = types.Output{
			OutputIndex: &outputIndex,
			Value:       &outputValue,
			ScriptHex:   rawVout.ScriptPubKey.Hex,
		}

		if len(rawVout.ScriptPubKey.Addresses) > 1 {
			// ScriptPubKey can have multiple addresses for multisig txns.
			//
			// Ref: https://bitcoin.stackexchange.com/a/4693/106367
			log.WithFields(log.Fields{
				"addresses":   rawVout.ScriptPubKey.Addresses,
				"value":       outputValue,
				"outputIndex": rawVout.N,
			}).Warnf("Multisig transaction detected.")
			vout[idx].Address = rawVout.ScriptPubKey.Addresses[0]
		} else if len(rawVout.ScriptPubKey.Addresses) == 1 {
			vout[idx].Address = rawVout.ScriptPubKey.Addresses[0]
		} else {
			// TODO: Document when this happens
		}

		sumVoutValues += *vout[idx].Value
	}
	txn.Outputs = vout

	txn.Block = block

	txn.Confirmations = rawTx.Confirmations

	if txn.Confirmations == 0 {
		// rawTx.Time is 0 if transaction is unconfirmed
		txn.ReceivedAt = utils.ParseUnixTimestamp(time.Now().Unix())
	} else {
		txn.ReceivedAt = utils.ParseUnixTimestamp(rawTx.Time)
	}

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
func (x XRPC) GetTransaction(txHash string) (*TransactionContainer, error) {
	txRaw, err := x.getTransactionByHash(txHash)
	if err != nil {
		return nil, err
	}

	utxos, err := x.buildUTXOs(txRaw.Vin)
	if err != nil {
		return nil, err
	}

	var block *types.Block
	if txRaw.BlockHash == "" {
		block = nil
	} else {
		block = &types.Block{
			Hash:   txRaw.BlockHash,
			Height: x.GetBlockHeightByHash(txRaw.BlockHash),
			Time:   utils.ParseUnixTimestamp(txRaw.Blocktime),
		}
	}

	transaction := new(TransactionContainer)
	transaction.init(txRaw, utxos, block)
	return transaction, nil
}

// GetTransactionHexByHash is a service function to get hex encoded raw
// transaction by hash.
func (x XRPC) GetTransactionHexByHash(txHash string) (string, error) {
	txRaw, err := x.getTransactionByHash(txHash)
	if err != nil {
		return "", err
	}
	return txRaw.Hex, nil
}

func (x XRPC) SendTransaction(tx string) (*string, error) {
	// Decode the serialized transaction hex to raw bytes.
	serializedTx, err := hex.DecodeString(tx)
	if err != nil {
		log.WithFields(log.Fields{
			"hex":   tx,
			"error": err,
		}).Error("Could not decode transaction hex")
		return nil, err
	}

	// Deserialize the transaction and return it.
	var msgTx wire.MsgTx
	if err := msgTx.Deserialize(bytes.NewReader(serializedTx)); err != nil {
		log.WithFields(log.Fields{
			"hex":   tx,
			"error": err,
		}).Error("Could not deserialize to wire.MsgTx")
		return nil, err
	}

	chainHash, err := x.SendRawTransaction(&msgTx, true)
	if err != nil {
		log.WithFields(log.Fields{
			"hex":   tx,
			"error": err,
		}).Error("sendrawtransaction RPC failed")
		return nil, err
	}

	txHash := chainHash.String()
	log.WithFields(log.Fields{
		"hex":  tx,
		"hash": txHash,
	}).Info("sendrawtransaction RPC successful")

	return btcjson.String(txHash), nil
}
