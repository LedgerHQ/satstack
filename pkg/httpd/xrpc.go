package httpd

import (
	"fmt"
	"ledger-sats-stack/pkg/transport"
	"ledger-sats-stack/pkg/utils"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	log "github.com/sirupsen/logrus"
)

// GetXRPC initializes an XRPC stuct that embeds a btcd RPC client.
func GetXRPC(host string, user string, pass string, tls bool) transport.XRPC {
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
	if client == nil || err != nil {
		log.WithFields(log.Fields{
			"host": host,
			"user": user,
			"TLS":  tls,
		}).Fatal("Failed to initialize RPC client")
	}

	info, err := client.GetBlockChainInfo()
	if info == nil || err != nil {
		log.WithFields(log.Fields{
			"host": host,
			"user": user,
			"TLS":  tls,
		}).Fatal("Failed to connect to RPC server")
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

	return transport.XRPC{
		Client:   client,
		Pruned:   info.Pruned,
		Chain:    info.Chain,
		TxIndex:  txIndex,
		Currency: getCurrencyFromChain(info.Chain),
	}
}

func isTxIndexEnabled(client *rpcclient.Client) bool {
	tx := getBlockOneTransaction(client)

	if _, err := client.GetRawTransaction(tx); err != nil {
		return false
	}

	return true
}

func getBlockOneTransaction(client *rpcclient.Client) *chainhash.Hash {
	// Genesis coinbase is not part of transaction index, so use block 1
	blockHash, err := client.GetBlockHash(1)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Failed to get block 1 hash")
	}

	block, err := client.GetBlockVerbose(blockHash)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Failed to get block 1")
	}

	coinbaseTxHash, err := utils.ParseChainHash(block.Tx[0])
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Failed to get coinbase tx in block 1")
	}

	return coinbaseTxHash
}

func WaitForNodeSync(xrpc transport.XRPC) {
	for {
		info, err := xrpc.Client.GetBlockChainInfo()
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
