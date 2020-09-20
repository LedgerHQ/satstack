package bus

import (
	"fmt"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/patrickmn/go-cache"
)

type Bus struct {
	Client   *rpcclient.Client
	Chain    string
	Pruned   bool
	TxIndex  bool
	Currency Currency // Based on Chain value, for interoperability with libcore
	Status   Status
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
		Status:   Initializing,
	}, nil
}

func (b *Bus) Close() {
	b.Client.Shutdown()
}

