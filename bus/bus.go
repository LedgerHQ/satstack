package bus

import (
	"fmt"
	"sync"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/patrickmn/go-cache"
)

// connPoolSize indicates the number of *rpcclient.Client objects that
// are available to use for communicating to the Bitcoin node.
//
// Set this to the maximum number of concurrent RPC operations that may be
// performed on the Bitcoin node.
const connPoolSize = 2

type Bus struct {
	// Informational fields
	Chain    string
	Pruned   bool
	TxIndex  bool
	Currency Currency // Based on Chain value, for interoperability with libcore
	Status   Status

	// Thread-safe Bus cache, to query results typically by hash
	Cache *cache.Cache

	// Connection pool management infrastructure
	connChan chan *rpcclient.Client
	wg       sync.WaitGroup
}

// New initializes a Bus struct that embeds a btcd RPC client.
func New(host string, user string, pass string, noTLS bool) (*Bus, error) {
	// Prepare the connection config to initialize the rpcclient.Client
	// pool with.
	connCfg := &rpcclient.ConnConfig{
		Host:         host,
		User:         user,
		Pass:         pass,
		HTTPPostMode: true,
		DisableTLS:   noTLS,
	}

	// Initialize a buffered channel of *rpcclient.Client objects with capacity
	// of connPoolSize.
	pool := make(chan *rpcclient.Client, connPoolSize)

	// Prefill the buffered channel with *rpcclient.Client objects in advance.
	for i := 0; i < cap(pool); i++ {
		client, err := rpcclient.New(connCfg, nil)
		if err != nil {
			return nil, err // error ctx not required
		}
		pool <- client
	}

	// Obtain one client from the channel to perform connectivity checks and
	// extract information required for initializing the Bus struct.
	client := <-pool
	defer func() { pool <- client }()

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

	b := &Bus{
		connChan: pool,
		Pruned:   info.Pruned,
		Chain:    info.Chain,
		TxIndex:  txIndex,
		Currency: currency,
		Cache:    nil, // Disabled by default
		Status:   Initializing,
		wg:       sync.WaitGroup{},
	}

	return b, nil
}

func (b *Bus) Close() {
	b.wg.Wait()

	for i := 0; i < cap(b.connChan); i++ {
		client := <-b.connChan
		client.WaitForShutdown()
		client.Shutdown()
	}

	close(b.connChan)
}

func (b *Bus) getClient() *rpcclient.Client {
	return <-b.connChan
}

func (b *Bus) recycleClient(client *rpcclient.Client) {
	b.connChan <- client
}
