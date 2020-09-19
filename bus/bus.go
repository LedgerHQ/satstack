package bus

import (
	"fmt"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
)

type Bus struct {
	Client   *rpcclient.Client
	Chain    string
	Pruned   bool
	TxIndex  bool
	Currency Currency     // Based on Chain value, for interoperability with libcore
	Cache    *cache.Cache // Thread-safe Bus cache, to query results typically by hash
}

// New initializes a Bus struct that embeds a btcd RPC client.
func New(host string, user string, pass string, noTLS bool) (*Bus, error) {
	connCfg := &rpcclient.ConnConfig{
		Host:         host,
		User:         user,
		Pass:         pass,
		HTTPPostMode: true,
		DisableTLS:   noTLS,
	}

	// The notification parameter is nil since notifications are not
	// supported in HTTP POST mode.
	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		return nil, err // error ctx not required
	}

	info, err := client.GetBlockChainInfo()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrBitcoindUnreachable, err)
	}

	txIndex, err := TxIndexEnabled(client)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrFailedToDetectTxIndex, err)
	}

	currency, err := CurrencyFromChain(info.Chain)
	if err != nil {
		return nil, err
	}

	return &Bus{
		Client:   client,
		Pruned:   info.Pruned,
		Chain:    info.Chain,
		TxIndex:  txIndex,
		Currency: currency,
		Cache:    nil, // Disabled by default
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
