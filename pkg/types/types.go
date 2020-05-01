package types

import (
	"github.com/btcsuite/btcutil"
)

// UTXO models the data corresponding to unspent transaction outputs.
// Convenience type; for limited use only.
type UTXO struct {
	Value   btcutil.Amount
	Address string
}

// Input models data corresponding to transaction inputs.
type Input struct {
	Coinbase    string          `json:"coinbase,omitempty"`         // [coinbase] The coinbase encoded as hex
	OutputHash  string          `json:"output_hash,omitempty"`      // [non-coinbase] Same as transaction ID of vin
	OutputIndex *uint32         `json:"output_index,omitempty"`     // [non-coinbase] Index of the corresponding UTXO
	Value       *btcutil.Amount `json:"value,omitempty"`            // [non-coinbase] Value of the corresponding UTXO in satoshis
	Address     string          `json:"address,omitempty"`          // [non-coinbase] Address of the corresponding UTXO; can be empty
	ScriptSig   *string         `json:"script_signature,omitempty"` // [non-coinbase] Hex-encoded signature script
	Witness     *[]string       `json:"txinwitness,omitempty"`      // [non-coinbase] Array of hex-encoded witness data
	InputIndex  *int            `json:"input_index,omitempty"`      // [all] Non-standard data required by Ledger Blockchain Explorer
	Sequence    uint32          `json:"sequence"`                   // [all] Input sequence number, used to track unconfirmed txns
}

// Output models data corresponding to transaction outputs.
type Output struct {
	OutputIndex *uint32         `json:"output_index,omitempty"` // Used to uniquely identify an output in a transaction
	Value       *btcutil.Amount `json:"value,omitempty"`        // Value of output in satoshis
	ScriptHex   string          `json:"script_hex"`             // Hex-encoded script
	Address     string          `json:"address,omitempty"`      // Address of the UTXO; can be empty
}

// Block models data corresponding to a block, but with limited information.
// It is used to represent minimal information of the block containing the given
// transaction.
type Block struct {
	Hash   string `json:"hash"`   // 0x prefixed
	Height int64  `json:"height"` // integer
	Time   string `json:"time"`   // RFC3339 format
}

// BlockWithTransactions is a struct that embeds Block, but also contains
// transaction hashes.
type BlockWithTransactions struct {
	Block
	Transactions []string `json:"txs"` // 0x prefixed
}

// Transaction represents the principal type to model the response of the GetTransaction handler.
type Transaction struct {
	ID            string          `json:"id"`
	Hash          string          `json:"hash"`
	ReceivedAt    string          `json:"received_at"`
	LockTime      uint32          `json:"lock_time"`
	Fees          *btcutil.Amount `json:"fees"`
	Confirmations uint64          `json:"confirmations"`
	Inputs        []Input         `json:"inputs"`
	Outputs       []Output        `json:"outputs"`
	Block         Block           `json:"block"`
}
