package controllers

import (
	"net/http"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gin-gonic/gin"
)

type utxoData struct {
	Value   int64
	Address string
}

type Input struct {
	Coinbase    string   `json:"coinbase,omitempty"`         // coinbase
	OutputHash  string   `json:"output_hash,omitempty"`      // non-coinbase
	OutputIndex uint32   `json:"output_index,omitempty"`     // non-coinbase
	Value       int64    `json:"value,omitempty"`            // non-coinbase
	Address     string   `json:"address,omitempty"`          // non-coinbase
	ScriptSig   string   `json:"script_signature,omitempty"` // non-coinbase
	TxInWitness []string `json:"txinwitness,omitempty"`      // non-coinbase
	InputIndex  int      `json:"input_index"`                // common
	Sequence    uint32   `json:"sequence"`                   // common
}

type Output struct {
	OutputIndex uint32 `json:"output_index"`      // common
	Value       int64  `json:"value"`             // common
	ScriptHex   string `json:"script_hex"`        // common
	Address     string `json:"address,omitempty"` // non-coinbase
}

type block struct {
	Hash   string `json:"hash"`
	Height int64  `json:"height"`
	Time   string `json:"time"`
}

type Transaction struct {
	ID            string   `json:"id"`
	Hash          string   `json:"hash"`
	ReceivedAt    string   `json:"received_at"`
	LockTime      uint32   `json:"lock_time"`
	Fees          int64    `json:"fees"`
	Confirmations uint64   `json:"confirmations"`
	Inputs        []Input  `json:"inputs"`
	Outputs       []Output `json:"outputs"`
	Block         block    `json:"block"`
}

func (txn *Transaction) init(rawTx *btcjson.TxRawResult, utxoMap map[string]map[uint32]utxoData, blockHeight int64) {
	txn.ID = rawTx.Txid
	txn.Hash = rawTx.Hash // Differs from ID for witness transactions
	txn.ReceivedAt = time.Unix(rawTx.Time, 0).Format(time.RFC3339)
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
				OutputHash:  rawVin.Txid, // Same as transaction ID of vin
				OutputIndex: rawVin.Vout, // UTXO index in the list of outputs of OutputHash
				InputIndex:  idx,         // TODO: Find out if the order matters
				Value:       utxo.Value,
				Address:     utxo.Address,
				ScriptSig:   rawVin.ScriptSig.Hex,
				Sequence:    rawVin.Sequence,
				TxInWitness: rawVin.Witness,
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
			Value:       int64(rawVout.Value * 100000000), // !FIXME: Can panic
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

	txn.Block = block{
		Hash:   rawTx.BlockHash,
		Height: blockHeight,
		Time:   time.Unix(rawTx.Blocktime, 0).Format(time.RFC3339),
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

// GetTransaction gets the transaction with the given hash.
// Supports transaction hashes with or without 0x prefix

func GetTransaction(client *rpcclient.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		transactionHashString := ctx.Param("hash")
		transactionHash, err := chainhash.NewHashFromStr(strings.TrimLeft(transactionHashString, "0x"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, err)
			return
		}

		txRaw, err := client.GetRawTransactionVerbose(transactionHash)
		if err != nil {
			ctx.JSON(http.StatusNotFound, err)
			return
		}

		utxoMap := make(map[string]map[uint32]utxoData)

		for _, inputRaw := range txRaw.Vin {
			if inputRaw.IsCoinBase() {
				continue
			}

			utxoHash, err := chainhash.NewHashFromStr(inputRaw.Txid)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, err)
				return
			}

			utxoTx, err := client.GetRawTransactionVerbose(utxoHash)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, err)
				return
			}
			utxoRaw := utxoTx.Vout[inputRaw.Vout]
			addresses := utxoRaw.ScriptPubKey.Addresses

			var utxo utxoData
			switch len(addresses) {
			case 0:
				// TODO: Document when this happens
				utxo = utxoData{
					int64(utxoRaw.Value * 100000000), // !FIXME: Can panic
					"",                               // Will be omitted by the JSON serializer
				}
			case 1:
				utxo = utxoData{
					int64(utxoRaw.Value * 100000000), // !FIXME: Can panic
					addresses[0],                     // ?XXX: Investigate why we do this
				}
			default:
				// TODO: Log an error
				utxo = utxoData{
					int64(utxoRaw.Value * 100000000), // !FIXME: Can panic
					"",                               // Will be omitted by the JSON serializer
				}
			}
			utxoMap[inputRaw.Txid] = make(map[uint32]utxoData)
			utxoMap[inputRaw.Txid][inputRaw.Vout] = utxo
		}

		blockHash, err := chainhash.NewHashFromStr(txRaw.BlockHash)
		if err != nil {
			// !FIXME: Handle the error appropriately
		}

		rawBlock, err := client.GetBlockVerbose(blockHash)
		var blockHeight int64
		if err != nil {
			blockHeight = -1
			// TODO: Log an error here
		} else {
			blockHeight = rawBlock.Height
		}

		transaction := new(Transaction)
		transaction.init(txRaw, utxoMap, blockHeight)

		ctx.JSON(http.StatusOK, []*Transaction{transaction})
	}
}
