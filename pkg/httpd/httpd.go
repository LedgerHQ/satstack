package httpd

import (
	"fmt"
	"ledger-sats-stack/pkg/handlers"
	"ledger-sats-stack/pkg/transport"
	"ledger-sats-stack/pkg/utils"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

func GetRouter(xrpc transport.XRPC, db *bolt.DB) *gin.Engine {
	engine := gin.Default()

	baseRouter := engine.Group("blockchain/v3")
	{
		baseRouter.GET("explorer/_health", handlers.GetHealth(xrpc))
		baseRouter.GET("syncToken", handlers.GetSyncToken(xrpc, db))
		baseRouter.DELETE("syncToken", handlers.DeleteSyncToken(xrpc, db))
	}

	var currency string
	info, _ := xrpc.GetBlockChainInfo()
	switch info.Chain {
	case "regtest", "test":
		currency = "btc_testnet"
	case "main":
		currency = "btc"
	}
	currencyRouter := baseRouter.Group(currency)
	{
		currencyRouter.GET("fees", handlers.GetFees(xrpc))
	}

	blocksRouter := currencyRouter.Group("/blocks")
	{
		blocksRouter.GET(":block", handlers.GetBlock(xrpc))
	}

	transactionsRouter := currencyRouter.Group("/transactions")
	{
		transactionsRouter.GET(":hash", handlers.GetTransaction(xrpc))
		transactionsRouter.GET(":hash/hex", handlers.GetTransactionHex(xrpc))
		transactionsRouter.POST("send", handlers.SendTransaction(xrpc))
	}

	addressesRouter := currencyRouter.Group("/addresses")
	{
		addressesRouter.GET(":addresses/transactions", handlers.GetAddresses(xrpc))
	}

	return engine
}

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

	waitForNodeSync(client)

	return transport.XRPC{
		Client:  client,
		Pruned:  info.Pruned,
		Chain:   info.Chain,
		TxIndex: txIndex,
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

func waitForNodeSync(client *rpcclient.Client) {
	for {
		info, err := client.GetBlockChainInfo()
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
			"progress": fmt.Sprintf("%.2f %%", info.VerificationProgress*100),
		}).Info("Sychronizing")

		time.Sleep(2 * time.Second)
	}
}
