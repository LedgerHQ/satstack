package bus

import (
	"errors"
	"fmt"
	"ledger-sats-stack/utils"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	log "github.com/sirupsen/logrus"
)

type Bus struct {
	Client   *rpcclient.Client
	Chain    string
	Pruned   bool
	TxIndex  bool
	Currency string // Based on Chain value, for interoperability with libcore
}

// New initializes a Bus struct that embeds a btcd RPC client.
func New(host string, user string, pass string, tls bool) (*Bus, error) {
	connCfg := &rpcclient.ConnConfig{
		Host:         host,
		User:         user,
		Pass:         pass,
		HTTPPostMode: true,
		DisableTLS:   !tls,
	}
	// The notification parameter is nil since notifications are not
	// supported in HTTP POST mode.
	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		log.WithFields(log.Fields{
			"host": host,
			"user": user,
			"TLS":  tls,
		}).Error("Failed to initialize RPC client")
		return nil, err
	}

	info, err := client.GetBlockChainInfo()
	if err != nil {
		log.WithFields(log.Fields{
			"host": host,
			"user": user,
			"TLS":  tls,
		}).Error("Failed to connect to RPC server")
		return nil, err
	}

	txIndex := isTxIndexEnabled(client)

	log.WithFields(log.Fields{
		"chain":   info.Chain,
		"pruned":  info.Pruned,
		"txindex": txIndex,
	}).Info("RPC connection established")

	if !txIndex {
		log.Warn("May have unexpected errors without txindex")
	}

	return &Bus{
		Client:   client,
		Pruned:   info.Pruned,
		Chain:    info.Chain,
		TxIndex:  txIndex,
		Currency: getCurrencyFromChain(info.Chain),
	}, nil
}

func (b *Bus) Close() {
	b.Client.Shutdown()
}

func WaitForNodeSync(bus *Bus) {
	for {
		info, err := bus.Client.GetBlockChainInfo()
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Fatal("Failed to connect to RPC server")
		}

		if info.Blocks == info.Headers {
			log.WithFields(log.Fields{
				"blocks":        info.Blocks,
				"bestblockhash": info.BestBlockHash,
			}).Info("Sychronization completed")
			return
		}

		log.WithFields(log.Fields{
			"progress": fmt.Sprintf("%.2f%%", info.VerificationProgress*100),
		}).Info("Sychronizing node")

		time.Sleep(10 * time.Second)
	}
}

func isTxIndexEnabled(client *rpcclient.Client) bool {
	tx, err := getBlockOneTransaction(client)
	if err != nil {
		log.Error(err)
		return false
	}

	if _, err := client.GetRawTransaction(tx); err != nil {
		return false
	}

	return true
}

func getBlockOneTransaction(client *rpcclient.Client) (*chainhash.Hash, error) {
	// Genesis coinbase is not part of transaction index, so use block 1
	blockHash, err := client.GetBlockHash(1)
	if err != nil {
		return nil, errors.New("failed to get block 1 hash")
	}

	block, err := client.GetBlockVerbose(blockHash)
	if err != nil {
		return nil, errors.New("dailed to get block 1")
	}

	coinbaseTxHash, err := utils.ParseChainHash(block.Tx[0])
	if err != nil {
		return nil, errors.New("failed to get coinbase tx in block 1")
	}

	return coinbaseTxHash, nil
}

func getCurrencyFromChain(chain string) string {
	switch chain {
	case "regtest", "test":
		return "btc_testnet"
	case "main":
		return "btc"
	default:
		return "btc"
	}
}
