package bus

import (
	"encoding/json"

	"fmt"

	"github.com/btcsuite/btcd/rpcclient"

	"github.com/ledgerhq/satstack/protocol"
	"github.com/ledgerhq/satstack/types"

	"github.com/ledgerhq/satstack/utils"

	"github.com/patrickmn/go-cache"

	"github.com/btcsuite/btcd/chaincfg/chainhash"

	"github.com/btcsuite/btcd/btcjson"
	log "github.com/sirupsen/logrus"
)

func (b *Bus) ListTransactions(blockHash *string) ([]btcjson.ListTransactionsResult, error) {
	var blockHashNative *chainhash.Hash
	if blockHash != nil {
		var err error
		blockHashNative, err = utils.ParseChainHash(*blockHash)
		if err != nil {
			return nil, err
		}
	}

	txs, err := b.mainClient.ListSinceBlockMinConfWatchOnly(blockHashNative, 1, true)
	if err != nil {
		return nil, err
	}

	return txs.Transactions, nil
}

func (b *Bus) GetTransactionHex(hash *chainhash.Hash) (string, error) {
	tx, err := b.mainClient.GetTransactionWatchOnly(hash, true)
	if err != nil {
		return "", err
	}

	return tx.Hex, nil
}

// see https://developer.bitcoin.org/reference/rpc/importdescriptors.html for specs
type ImportDesciptorRequest struct {
	Descriptor string `json:"desc"`                 //(string, required) Descriptor to import.
	Active     bool   `json:"active,omitempty"`     //(boolean, optional, default=false) Set this descriptor to be the active descriptor for the corresponding output type/externality
	Range      []int  `json:"range,omitempty"`      //(numeric or array) If a ranged descriptor is used, this specifies the end or the range (in the form [begin,end]) to import
	NextIndex  int    `json:"next_index,omitempty"` //(numeric) If a ranged descriptor is set to active, this specifies the next index to generate addresses from
	Timestamp  uint32 `json:"timestamp"`            /*(integer / string, required) Time from which to start rescanning the blockchain for this descriptor, in UNIX epoch time
	Use the string "now" to substitute the current synced blockchain time.
	"now" can be specified to bypass scanning, for outputs which are known to never have been used, and
	0 can be specified to scan the entire blockchain. Blocks up to 2 hours before the earliest timestamp
	of all descriptors being imported will be scanned.*/
	Internal bool `json:"internal,omitempty"` //(boolean, optional, default=false) Whether matching outputs should be treated as not incoming payments (e.g. change)
	// Label    string `json:"label",omitempty`    //(string, optional, default='') Label to assign to the address, only allowed with internal=false
}

type ImportDescriptorResult struct {
	Success  bool             `json:"success"`
	Warnings []string         `json:"warnings"`
	Error    btcjson.RPCError `json:"error"`
}

func ImportDescriptors(client *rpcclient.Client, descriptors []descriptor) error {

	// We are going to import all descriptors together which saves us a lot of time

	var requestDescriptors []ImportDesciptorRequest
	var params []json.RawMessage

	for _, descriptor := range descriptors {

		requests := ImportDesciptorRequest{
			Descriptor: descriptor.Value,
			Active:     true,
			Range:      []int{0, descriptor.Depth},
			Timestamp:  descriptor.Age,
		}

		requestDescriptors = append(requestDescriptors, requests)

	}

	myIn, mErr := json.Marshal(requestDescriptors)

	if mErr != nil {
		log.Error(`mErr`, mErr)
		return mErr
	}

	myInRaw := json.RawMessage(myIn)
	params = append(params, myInRaw)

	method := "importdescriptors"

	result, err := client.RawRequest(method, params)

	if err != nil {
		log.Error(`err `, err)
		return err
	}

	var importDescriptorResult []ImportDescriptorResult
	umerr := json.Unmarshal(result, &importDescriptorResult)

	if umerr != nil {
		log.Error(`umerr `, umerr)
		return umerr
	}

	var hasError bool

	fields := log.WithFields(log.Fields{
		"NumofDescriptors": len(requestDescriptors),
	})

	if !importDescriptorResult[0].Success {

		fields.Error("ImportDescriptors - Failed to import descriptor" + " || " + importDescriptorResult[0].Error.Error())
		hasError = true
	} else {
		fields.Debug("ImportDescriptors - Import descriptor successfully")
	}

	if hasError {
		return fmt.Errorf("ImportDescriptors - importdescriptor RPC failed")
	}

	return nil

}

func (b *Bus) GetTransaction(hash string) (*types.Transaction, error) {
	if b.Cache != nil { // Cache has been enabled at the svc level
		if tx, found := b.Cache.Get(hash); found {
			return tx.(*types.Transaction), nil
		}
	}

	chainHash, err := utils.ParseChainHash(hash)
	if err != nil {
		return nil, err
	}

	var tx *types.Transaction

	switch b.TxIndex {
	case true:
		txRaw, err := b.mainClient.GetRawTransaction(chainHash)
		if err != nil {
			return nil, err
		}

		tx = protocol.DecodeMsgTx(txRaw.MsgTx(), b.Params)

	case false:
		txRaw, err := b.mainClient.GetTransactionWatchOnly(chainHash, true)
		if err != nil {
			return nil, err
		}

		tx, err = protocol.DecodeRawTransaction(txRaw.Hex, b.Params)
		if err != nil {
			return nil, err
		}
	}

	if b.Cache != nil {
		b.Cache.Set(hash, tx, cache.NoExpiration)
	}

	return tx, nil
}
