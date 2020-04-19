package controllers

import (
	"net/http"
	"strings"

	"ledger-sats-stack/httpd/utils"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gin-gonic/gin"
)

// UTXO models the data corresponding to unspent transaction outputs.
// Convenience type; for limited use only.
type UTXO struct {
	Value   int64
	Address string
}

// Input models data corresponding to transaction inputs.
type Input struct {
	Coinbase    string   `json:"coinbase,omitempty"`         // [coinbase] The coinbase encoded as hex
	OutputHash  string   `json:"output_hash,omitempty"`      // [non-coinbase] Same as transaction ID of vin
	OutputIndex uint32   `json:"output_index,omitempty"`     // [non-coinbase] Index of the corresponding UTXO
	Value       int64    `json:"value,omitempty"`            // [non-coinbase] Value of the corresponding UTXO in satoshis
	Address     string   `json:"address,omitempty"`          // [non-coinbase] Address of the corresponding UTXO; can be empty
	ScriptSig   string   `json:"script_signature,omitempty"` // [non-coinbase] Hex-encoded signature script
	Witness     []string `json:"txinwitness,omitempty"`      // [non-coinbase] Array of hex-encoded witness data
	InputIndex  int      `json:"input_index"`                // [all] Non-standard data required by Ledger Blockchain Explorer
	Sequence    uint32   `json:"sequence"`                   // [all] Input sequence number, used to track unconfirmed txns
}

// Output models data corresponding to transaction outputs.
type Output struct {
	OutputIndex uint32 `json:"output_index"`      // Used to uniquely identify an output in a transaction
	Value       int64  `json:"value"`             // Value of output in satoshis
	ScriptHex   string `json:"script_hex"`        // Hex-encoded script
	Address     string `json:"address,omitempty"` // Address of the UTXO; can be empty
}

// SparseBlock models data corresponding to a block, but with limited information.
// It is used to represent minimal information of the block containing the given
// transaction.
type SparseBlock struct {
	Hash   string `json:"hash"`
	Height int64  `json:"height"`
	Time   string `json:"time"`
}

// Transaction represents the principal type to model the response of the GetTransaction handler.
type Transaction struct {
	ID            string      `json:"id"`
	Hash          string      `json:"hash"`
	ReceivedAt    string      `json:"received_at"`
	LockTime      uint32      `json:"lock_time"`
	Fees          int64       `json:"fees"`
	Confirmations uint64      `json:"confirmations"`
	Inputs        []Input     `json:"inputs"`
	Outputs       []Output    `json:"outputs"`
	Block         SparseBlock `json:"block"`
}

func (txn *Transaction) init(rawTx *btcjson.TxRawResult, utxoMap map[string]map[uint32]UTXO, blockHeight int64) {
	txn.ID = rawTx.Txid
	txn.Hash = rawTx.Hash // Differs from ID for witness transactions
	txn.ReceivedAt = utils.ParseUnixTimestamp(rawTx.Time)
	txn.LockTime = rawTx.LockTime

	vin := make([]Input, len(rawTx.Vin))
	sumVinValues := int64(0)
	vinHasCoinbase := false

	for idx, rawVin := range rawTx.Vin {
		if rawVin.IsCoinBase() {
			vin[idx] = Input{
				Coinbase:   rawVin.Coinbase,
				InputIndex: idx,
				Sequence:   rawVin.Sequence,
			}

			vinHasCoinbase = true
		} else {
			utxo := utxoMap[rawVin.Txid][rawVin.Vout]
			vin[idx] = Input{
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

	vout := make([]Output, len(rawTx.Vout))
	sumVoutValues := int64(0)

	for idx, rawVout := range rawTx.Vout {
		vout[idx] = Output{
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

	txn.Block = SparseBlock{
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

// GetTransaction is a gin handler (factory) to query transaction details
// by hash parameter.
func GetTransaction(client *rpcclient.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		txHash := ctx.Param("hash")
		txRaw, err := getTransactionByHash(client, txHash)
		if err != nil {
			ctx.JSON(http.StatusNotFound, err)
			return
		}

		utxoMap, err := buildUtxoMap(client, txRaw.Vin)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, err)
			return
		}

		blockHeight := getBlockHeightByHash(client, txRaw.BlockHash)

		transaction := new(Transaction)
		transaction.init(txRaw, utxoMap, blockHeight)

		ctx.JSON(http.StatusOK, []*Transaction{transaction})
	}
}

func buildUtxoMap(client *rpcclient.Client, vin []btcjson.Vin) (map[string]map[uint32]UTXO, error) {
	utxoMap := make(map[string]map[uint32]UTXO)

	for _, inputRaw := range vin {
		if inputRaw.IsCoinBase() {
			continue
		}

		txn, err := getTransactionByHash(client, inputRaw.Txid)
		if err != nil {
			return nil, err
		}
		utxoRaw := txn.Vout[inputRaw.Vout]
		addresses := utxoRaw.ScriptPubKey.Addresses

		var utxo UTXO
		switch len(addresses) {
		case 0:
			// TODO: Document when this happens
			utxo = UTXO{
				utils.ParseSatoshi(utxoRaw.Value), // !FIXME: Can panic
				"",                                // Will be omitted by the JSON serializer
			}
		case 1:
			utxo = UTXO{
				utils.ParseSatoshi(utxoRaw.Value), // !FIXME: Can panic
				addresses[0],                      // ?XXX: Investigate why we do this
			}
		default:
			// TODO: Log an error
			utxo = UTXO{
				utils.ParseSatoshi(utxoRaw.Value), // !FIXME: Can panic
				"",                                // Will be omitted by the JSON serializer
			}
		}
		utxoMap[inputRaw.Txid] = make(map[uint32]UTXO)
		utxoMap[inputRaw.Txid][inputRaw.Vout] = utxo
	}

	return utxoMap, nil
}

func getBlockHeightByHash(client *rpcclient.Client, hash string) int64 {
	hashRaw, err := chainhash.NewHashFromStr(hash)
	if err != nil {
		// TODO: Log an error here
		return -1
	}

	rawBlock, err := client.GetBlockVerbose(hashRaw)

	if err != nil {
		// TODO: Log an error here
		return -1

	}
	return rawBlock.Height
}

// getTransactionByHash gets the transaction with the given hash.
// Supports transaction hashes with or without 0x prefix.
func getTransactionByHash(client *rpcclient.Client, hash string) (*btcjson.TxRawResult, error) {
	txHashRaw, err := chainhash.NewHashFromStr(strings.TrimLeft(hash, "0x"))
	if err != nil {
		return nil, err
	}

	txRaw, err := client.GetRawTransactionVerbose(txHashRaw)
	if err != nil {
		return nil, err
	}
	return txRaw, nil
}
