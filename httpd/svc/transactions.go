package svc

import (
	"time"

	"github.com/ledgerhq/satstack/types"
	"github.com/ledgerhq/satstack/utils"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcutil"
	log "github.com/sirupsen/logrus"
)

// GetTransaction is a service function to query transaction details
// by transaction hash.
func (s *Service) GetTransaction(hash string, block *types.Block, bestBlockHeight int32) (*types.Transaction, error) {
	chainHash, err := utils.ParseChainHash(hash)
	if err != nil {
		return nil, err
	}

	tx, err := s.Bus.GetTransaction(chainHash)
	if err != nil {
		return nil, err
	}

	utxos, err := s.buildUTXOs(tx.Inputs)
	if err != nil {
		return nil, err
	}

	tx.Block = block
	buildTx(tx, utxos, bestBlockHeight)

	return tx, nil
}

// GetTransactionHex is a service function to get hex encoded raw
// transaction by hash.
func (s *Service) GetTransactionHex(hash string) (string, error) {
	chainHash, err := utils.ParseChainHash(hash)
	if err != nil {
		return "", err
	}

	return s.Bus.GetTransactionHex(chainHash)
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

func (s *Service) buildUTXOs(vin []types.Input) (types.UTXOs, error) {
	utxoMap := make(types.UTXOs)

	for _, inputRaw := range vin {
		if len(inputRaw.Coinbase) > 0 {
			continue
		}

		utxoID := types.OutputIdentifier{
			Hash:  inputRaw.OutputHash,
			Index: *inputRaw.OutputIndex, // FIXME: can panic
		}

		utxoHash, err := utils.ParseChainHash(utxoID.Hash)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"hash":  utxoID.Hash,
				"vout":  utxoID.Index,
			}).Error("Could not parse UTXO hash")
			continue
		}

		utxo, err := s.Bus.GetTransaction(utxoHash)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"hash":  utxoID.Hash,
				"vout":  utxoID.Index,
			}).Debug("Encountered non-wallet Vout")
			continue
		}

		utxoMap[utxoID] = types.UTXOData{
			Value:   *utxo.Outputs[utxoID.Index].Value, // FIXME: can panic
			Address: utxo.Outputs[utxoID.Index].Address,
		}
	}

	return utxoMap, nil
}

func buildTx(tx *types.Transaction, utxoMap types.UTXOs, bestBlockHeight int32) {
	sumVinValues := btcutil.Amount(0)
	vinHasCoinbase := false

	for idx, vin := range tx.Inputs {
		if len(vin.Coinbase) > 0 {
			vinHasCoinbase = true
			continue
		}

		utxoID := types.OutputIdentifier{
			Hash:  vin.OutputHash,
			Index: *vin.OutputIndex,
		}

		utxo := utxoMap[utxoID]

		tx.Inputs[idx].Address = utxo.Address // mutate the vins in tx
		tx.Inputs[idx].Value = &utxo.Value

		sumVinValues += utxo.Value
	}

	sumVoutValues := btcutil.Amount(0)

	for _, vout := range tx.Outputs {
		sumVoutValues += *vout.Value
	}

	if tx.Block != nil {
		tx.Confirmations = uint64(int64(bestBlockHeight)-tx.Block.Height) + 1
		tx.ReceivedAt = tx.Block.Time
	} else {
		// Handle the case of unconfirmed transaction.
		tx.Confirmations = 0
		tx.ReceivedAt = utils.ParseUnixTimestamp(time.Now().Unix())
	}

	var fees btcutil.Amount

	if vinHasCoinbase {
		// Coinbase transactions have no fees
		fees = btcutil.Amount(0)
	} else {
		fees = sumVinValues - sumVoutValues
	}

	// This is typically in case of incoming transactions, where computing
	// the fees without a transaction index is impossible.
	if fees < 0 {
		fees = 0
	}

	tx.Fees = &fees

	// In Ledger Blockchain Explorer v2, the Amount field is the sum of all
	// Vout values.
	tx.Amount = &sumVoutValues
}
