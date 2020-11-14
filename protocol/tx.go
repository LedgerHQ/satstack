package protocol

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/ledgerhq/satstack/types"
)

func DecodeRawTransaction(txnHex string, params *chaincfg.Params) (*types.Transaction, error) {
	hexStr := txnHex

	// Left-pad with zero if length of transaction hex is not even.
	if len(hexStr)%2 != 0 {
		hexStr = "0" + hexStr
	}

	serializedTx, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("%s: %w: %s", ErrDecodeHex, err, txnHex)
	}

	var mtx wire.MsgTx

	err = mtx.Deserialize(bytes.NewReader(serializedTx))
	if err != nil {
		return nil, fmt.Errorf("%s: %w: %s", ErrMsgTxDeserialize, err, txnHex)
	}

	return &types.Transaction{
		ID:       mtx.TxHash().String(),
		Hash:     mtx.TxHash().String(),
		LockTime: mtx.LockTime,
		Inputs:   createVinList(&mtx),
		Outputs:  createVoutList(&mtx, params),
	}, nil
}

// createVinList returns a slice of JSON objects for the inputs of the passed
// transaction.
func createVinList(mtx *wire.MsgTx) []types.Input {
	// Coinbase transactions only have a single TxIn by definition.
	vinList := make([]types.Input, len(mtx.TxIn))
	if blockchain.IsCoinBaseTx(mtx) {
		txIn := mtx.TxIn[0]

		vinList[0].InputIndex = btcjson.Int(0)
		vinList[0].Coinbase = hex.EncodeToString(txIn.SignatureScript)
		vinList[0].Sequence = txIn.Sequence
		vinList[0].Witness = witnessToHex(txIn.Witness)
		return vinList
	}

	for i, txIn := range mtx.TxIn {
		vinEntry := &vinList[i]
		vinEntry.InputIndex = btcjson.Int(i)
		vinEntry.OutputHash = txIn.PreviousOutPoint.Hash.String()
		vinEntry.OutputIndex = btcjson.Uint32(txIn.PreviousOutPoint.Index)
		vinEntry.Sequence = txIn.Sequence
		vinEntry.ScriptSig = btcjson.String(
			hex.EncodeToString(txIn.SignatureScript))

		if mtx.HasWitness() {
			vinEntry.Witness = witnessToHex(txIn.Witness)
		}
	}

	return vinList
}

// createVoutList returns a slice of JSON objects for the outputs of the passed
// transaction.
func createVoutList(mtx *wire.MsgTx, chainParams *chaincfg.Params) []types.Output {
	voutList := make([]types.Output, 0, len(mtx.TxOut))
	for i, v := range mtx.TxOut {
		var vout types.Output

		vout.OutputIndex = btcjson.Uint32(uint32(i))
		value := btcutil.Amount(v.Value)
		vout.Value = &value
		vout.ScriptHex = hex.EncodeToString(v.PkScript)

		// Ignore the error here since an error means the script
		// couldn't parse. In such a case, addrs will be nil.
		_, addrs, _, _ := txscript.ExtractPkScriptAddrs(
			v.PkScript, chainParams)

		// Encode the addresses to to string.
		encodedAddrs := make([]string, len(addrs))
		for j, addr := range addrs {
			encodedAddr := addr.EncodeAddress()
			encodedAddrs[j] = encodedAddr
		}

		// ScriptPubKey can have multiple addresses for multisig transactions.
		//
		// We pick the first address in the list, which is what libcore
		// expects. Caution: may have side-effects.
		//
		// In case of no addresses, the Address field is not populated.
		// Generally, this means the ScriptPubKey is corrupt.
		//
		// Ref: https://bitcoin.stackexchange.com/a/4693/106367
		if len(encodedAddrs) > 0 {
			vout.Address = encodedAddrs[0]
		}

		voutList = append(voutList, vout)
	}

	return voutList
}

// witnessToHex formats the passed witness stack as a slice of hex-encoded
// strings to be used in a JSON response.
func witnessToHex(witness wire.TxWitness) []string {
	// Ensure nil is returned when there are no entries versus an empty
	// slice so it can properly be omitted as necessary.
	if len(witness) == 0 {
		return nil
	}

	result := make([]string, 0, len(witness))
	for _, wit := range witness {
		result = append(result, hex.EncodeToString(wit))
	}

	return result
}
