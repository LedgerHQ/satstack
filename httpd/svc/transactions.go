package svc

import (
	"time"

	"ledger-sats-stack/types"
	"ledger-sats-stack/utils"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcutil"
	log "github.com/sirupsen/logrus"
)

// GetTransaction is a service function to query transaction details
// by transaction hash.
func (s *Service) GetTransaction(hash string, block *types.Block) (*types.Transaction, error) {
	chainHash, err := utils.ParseChainHash(hash)
	if err != nil {
		return nil, err
	}

	tx, err := s.Bus.GetTransaction(chainHash)
	if err != nil {
		return nil, err
	}

	utxos, err := s.buildUTXOs(tx.Vin)
	if err != nil {
		return nil, err
	}

	transaction := buildTx(tx, utxos)

	if block == nil {
		transaction.Block = s.getTransactionBlock(tx)
	} else {
		transaction.Block = block
	}

	return transaction, nil
}

// GetTransactionHex is a service function to get hex encoded raw
// transaction by hash.
func (s *Service) GetTransactionHex(hash string) (*string, error) {
	chainHash, err := utils.ParseChainHash(hash)
	if err != nil {
		return nil, err
	}

	tx, err := s.Bus.GetTransaction(chainHash)
	if err != nil {
		return nil, err
	}
	return utils.ToStringPointer(tx.Hex), nil
}

func (s *Service) SendTransaction(tx string) (*string, error) {
	hash, err := s.Bus.SendTransaction(tx)
	if err != nil {
		return nil, err
	}
	return utils.ToStringPointer(hash.String()), nil
}

func (s *Service) getTransactionBlock(tx *btcjson.TxRawResult) *types.Block {
	if tx.BlockHash == "" {
		return nil
	}

	chainHash, err := utils.ParseChainHash(tx.BlockHash)
	if err != nil {
		return nil
	}

	block, err := s.Bus.GetBlock(chainHash)
	if err != nil {
		return nil
	}

	return &types.Block{
		Hash:   tx.BlockHash,
		Height: block.Height,
		Time:   utils.ParseUnixTimestamp(tx.Blocktime),
	}
}

func (s *Service) buildUTXOs(vin []btcjson.Vin) (types.UTXOs, error) {
	utxos := make(types.UTXOs)
	utxoResults := make(map[types.OutputIdentifier]*btcjson.TxRawResult)

	for _, inputRaw := range vin {
		if inputRaw.IsCoinBase() {
			continue
		}

		utxoHash, err := utils.ParseChainHash(inputRaw.Txid)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"hash":  inputRaw.Txid,
				"vout":  inputRaw.Vout,
			}).Error("Could not parse UTXO hash")
			continue
		}

		utxo, err := s.Bus.GetTransaction(utxoHash)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"hash":  inputRaw.Txid,
				"vout":  inputRaw.Vout,
			}).Warn("Encountered non-wallet Vout")
			continue
		}

		utxoResults[types.OutputIdentifier{Hash: inputRaw.Txid, Index: inputRaw.Vout}] = utxo
	}

	for utxoID, utxoResult := range utxoResults {
		utxo, err := parseUTXO(utxoResult, utxoID.Index)
		if err != nil {
			return nil, err
		}

		utxos[utxoID] = *utxo
	}

	return utxos, nil
}

func parseUTXO(tx *btcjson.TxRawResult, outputIndex uint32) (*types.UTXOData, error) {
	utxoRaw := tx.Vout[outputIndex]

	switch addresses := utxoRaw.ScriptPubKey.Addresses; len(addresses) {
	case 0:
		// TODO: Document when this happens
		return &types.UTXOData{
			Value:   utils.ParseSatoshi(utxoRaw.Value), // !FIXME: Can panic
			Address: "",                                // Will be omitted by the JSON serializer
		}, nil
	case 1:
		return &types.UTXOData{
			Value:   utils.ParseSatoshi(utxoRaw.Value),
			Address: addresses[0], // ?XXX: Investigate why we do this
		}, nil
	default:
		value := utils.ParseSatoshi(utxoRaw.Value) // !FIXME: Can panic
		log.WithFields(log.Fields{
			"addresses":   addresses,
			"value":       value,
			"outputIndex": outputIndex,
		}).Warn("Multisig transaction detected.")

		return &types.UTXOData{
			Value:   value,
			Address: addresses[0],
		}, nil
	}
}

func buildTx(rawTx *btcjson.TxRawResult, utxos types.UTXOs) *types.Transaction {
	txn := new(types.Transaction)
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
			log.WithFields(log.Fields{
				"value":       outputValue,
				"outputIndex": rawVout.N,
			}).Warn("No address in scriptPubKey")
		}

		sumVoutValues += *vout[idx].Value
	}
	txn.Outputs = vout

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

	return txn
}
