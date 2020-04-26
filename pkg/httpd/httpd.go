package httpd

import (
	"ledger-sats-stack/pkg/handlers"
	"ledger-sats-stack/pkg/transport"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func GetRouter(wire transport.Wire) *gin.Engine {
	engine := gin.Default()

	engine.GET("/blockchain/v3/blocks/:block", handlers.GetBlock(wire))

	engine.GET("/blockchain/v3/transactions/:hash", handlers.GetTransaction(wire))
	engine.GET("/blockchain/v3/transactions/:hash/hex", handlers.GetTransactionHex(wire))

	engine.GET("/blockchain/v3/explorer/_health", handlers.GetHealth(wire))

	return engine
}

// GetWire initializes a Wire stuct that embeds an RPC client.
func GetWire(host string, user string, pass string, tls bool) transport.Wire {
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
		}).Fatal("Failed to initialize RPC client.")

	}

	info, err := client.GetBlockChainInfo()

	if err != nil {
		log.WithFields(log.Fields{
			"host": host,
			"user": user,
			"TLS":  tls,
		}).Fatal("Failed to connect to RPC server.")
	}

	log.WithFields(log.Fields{
		"chain":         info.Chain,
		"blocks":        info.Blocks,
		"bestblockhash": info.BestBlockHash,
	}).Info("RPC connection established.")

	return transport.Wire{Client: client}
}
