package transactions

import (
	"strings"

	blocksRPC "ledger-sats-stack/app/rpc/blocks"
	"ledger-sats-stack/app/types"
	"ledger-sats-stack/app/utils"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
)

type utxoVoutMapType map[uint32]types.UTXO
type utxoMapType map[string]utxoVoutMapType

// TransactionContainer is a wrapper type to define an init method for
// Transaction
type TransactionContainer struct {
	types.Transaction
}

func (txn *TransactionContainer) init(rawTx *btcjson.TxRawResult, utxoMap utxoMapType, blockHeight int64) {
	txn.ID = rawTx.Txid
	txn.Hash = rawTx.Hash // Differs from ID for witness transactions
	txn.ReceivedAt = utils.ParseUnixTimestamp(rawTx.Time)
	txn.LockTime = rawTx.LockTime

	vin := make([]types.Input, len(rawTx.Vin))
	sumVinValues := int64(0)
	vinHasCoinbase := false

	for idx, rawVin := range rawTx.Vin {
		if rawVin.IsCoinBase() {
			vin[idx] = types.Input{
				Coinbase:   rawVin.Coinbase,
				InputIndex: idx,
				Sequence:   rawVin.Sequence,
			}

			vinHasCoinbase = true
		} else {
			utxo := utxoMap[rawVin.Txid][rawVin.Vout]
			vin[idx] = types.Input{
				OutputHash:  rawVin.Txid,
				OutputIndex: rawVin.Vout,
				InputIndex:  idx, // TODO: Find out if the order matters
				Value:       utxo.Value,
				Address:     utxo.Address,
				ScriptSig:   rawVin.ScriptSig.Hex,
				Sequence:    rawVin.Sequence,
				Witness:     rawVin.Witness, // !FIXME: Coinbase txn can also have witness
			}

			sumVinValues += vin[idx].Value
		}
	}
	txn.Inputs = vin

	vout := make([]types.Output, len(rawTx.Vout))
	sumVoutValues := int64(0)

	for idx, rawVout := range rawTx.Vout {
		vout[idx] = types.Output{
			OutputIndex: rawVout.N,
			Value:       utils.ParseSatoshi(rawVout.Value), // !FIXME: Can panic
			ScriptHex:   rawVout.ScriptPubKey.Hex,
		}

		if len(rawVout.ScriptPubKey.Addresses) == 1 {
			vout[idx].Address = rawVout.ScriptPubKey.Addresses[0]
		} else if len(rawVout.ScriptPubKey.Addresses) > 1 {
			// TODO: Log an error / warning
		} else {
			// TODO: Document when this happens
		}

		sumVoutValues += vout[idx].Value
	}
	txn.Outputs = vout

	txn.Block = types.Block{
		Hash:   rawTx.BlockHash,
		Height: blockHeight,
		Time:   utils.ParseUnixTimestamp(rawTx.Blocktime),
	}

	// ?XXX: Confirmations in Ledger Blockchain Explorer are always off by 1
	txn.Confirmations = rawTx.Confirmations - uint64(1)

	if vinHasCoinbase {
		// Coinbase transaction have no fees
		txn.Fees = int64(0)
	} else {
		txn.Fees = sumVinValues - sumVoutValues
	}
}

func buildUtxoMap(client *rpcclient.Client, vin []btcjson.Vin) (utxoMapType, error) {
	utxoMap := make(utxoMapType)

	for _, inputRaw := range vin {
		if inputRaw.IsCoinBase() {
			continue
		}

		txn, err := getTransactionByHash(client, inputRaw.Txid)
		if err != nil {
			return nil, err
		}
		utxoRaw := txn.Vout[inputRaw.Vout]

		utxo := func(addresses []string) types.UTXO {
			switch len(addresses) {
			case 0:
				// TODO: Document when this happens
				return types.UTXO{
					Value:   utils.ParseSatoshi(utxoRaw.Value), // !FIXME: Can panic
					Address: "",                                // Will be omitted by the JSON serializer
				}
			case 1:
				return types.UTXO{
					Value:   utils.ParseSatoshi(utxoRaw.Value),
					Address: addresses[0], // ?XXX: Investigate why we do this
				}
			default:
				// TODO: Log an error
				return types.UTXO{
					Value:   utils.ParseSatoshi(utxoRaw.Value), // !FIXME: Can panic
					Address: "",                                // Will be omitted by the JSON serializer
				}
			}
		}(utxoRaw.ScriptPubKey.Addresses)

		utxoMap[inputRaw.Txid] = make(utxoVoutMapType)
		utxoMap[inputRaw.Txid][inputRaw.Vout] = utxo
	}

	return utxoMap, nil
}

// getTransactionByHash gets the transaction with the given hash.
// Supports transaction hashes with or without 0x prefix.
func getTransactionByHash(client *rpcclient.Client, txHash string) (*btcjson.TxRawResult, error) {
	txHashRaw, err := chainhash.NewHashFromStr(strings.TrimLeft(txHash, "0x"))
	if err != nil {
		return nil, err
	}

	txRaw, err := client.GetRawTransactionVerbose(txHashRaw)
	if err != nil {
		return nil, err
	}
	return txRaw, nil
}

// GetTransaction is a service function to query transaction details
// by hash parameter.
func GetTransaction(txHash string, client *rpcclient.Client) (*TransactionContainer, error) {
	txRaw, err := getTransactionByHash(client, txHash)
	if err != nil {
		return nil, err
	}

	utxoMap, err := buildUtxoMap(client, txRaw.Vin)
	if err != nil {
		return nil, err
	}

	blockHeight := blocksRPC.GetBlockHeightByHash(client, txRaw.BlockHash)

	transaction := new(TransactionContainer)
	transaction.init(txRaw, utxoMap, blockHeight)
	return transaction, nil
}

// GetTransactionHexByHash is a service function to get hex encoded raw
// transaction by hash.
func GetTransactionHexByHash(client *rpcclient.Client, txHash string) (string, error) {
	txRaw, err := getTransactionByHash(client, txHash)
	if err != nil {
		return "", err
	}
	return txRaw.Hex, nil
}
