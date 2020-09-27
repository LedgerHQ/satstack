package bus

import (
	"bytes"
	"encoding/hex"

	"github.com/btcsuite/btcd/chaincfg/chainhash"

	"github.com/btcsuite/btcd/wire"
	log "github.com/sirupsen/logrus"
)

func (b *Bus) SendTransaction(tx string) (*chainhash.Hash, error) {
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

	client := <-b.connChan
	defer func() { b.connChan <- client }()

	chainHash, err := client.SendRawTransaction(&msgTx, true)
	if err != nil {
		log.WithFields(log.Fields{
			"hex":   tx,
			"error": err,
		}).Error("sendrawtransaction Bridge failed")
		return nil, err
	}

	log.WithFields(log.Fields{
		"hex":  tx,
		"hash": chainHash.String(),
	}).Info("sendrawtransaction Bridge successful")

	return chainHash, nil
}
