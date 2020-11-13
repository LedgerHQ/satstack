package bus

import (
	"fmt"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcutil"
	"github.com/ledgerhq/satstack/config"
	"github.com/ledgerhq/satstack/utils"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
)

const (
	// defaultAccountDepth indicates the number of addresses to derive and
	// import in the Bitcoin wallet.
	defaultAccountDepth = 1000

	// connPoolSize indicates the number of *rpcclient.Client objects that
	// are available to use for communicating to the Bitcoin node.
	//
	// Set this to the maximum number of concurrent RPC operations that may be
	// performed on the Bitcoin node.
	connPoolSize = 2

	// minimumSupportedBitcoindVersion indicates the minimum version that is
	// supported by SatStack.
	minSupportedBitcoindVersion = 200000

	// walletName indicates the name of the wallet created by SatStack in
	// bitcoind's wallet.
	walletName = "satstack"

	errDuplicateWalletLoadMsg = "Duplicate -wallet filename specified."
)

// Bus represents a transport allowing access to Bitcoin RPC methods.
//
// It maintains a pool of btcd rpcclient objects in a buffered channel to allow
// concurrent invocation of RPC methods.
type Bus struct {
	// Informational fields
	Chain       string
	Pruned      bool
	TxIndex     bool
	BlockFilter bool
	Currency    Currency // Based on Chain value, for interoperability with libcore
	Status      Status

	// Thread-safe Bus cache, to query results typically by hash
	Cache *cache.Cache

	// Connection pool management infrastructure
	connChan chan *rpcclient.Client
}

type descriptor struct {
	Value string
	Depth int
	Age   uint32
}

// New initializes a Bus struct that embeds a btcd RPC client.
func New(host string, user string, pass string, noTLS bool) (*Bus, error) {
	// Prepare the connection config to initialize the rpcclient.Client
	// pool with.
	connCfg := &rpcclient.ConnConfig{
		Host:         fmt.Sprintf("%s/wallet/%s", host, walletName),
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

	networkInfo, err := client.GetNetworkInfo()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrBitcoindUnreachable, err)
	}

	if v := networkInfo.Version; v < minSupportedBitcoindVersion {
		return nil, fmt.Errorf("%s: %d", ErrUnsupportedBitcoindVersion, v)
	}

	blockFilter, err := blockFilterEnabled(client, info.BestBlockHash)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrFailedToDetectBlockFilter, err)
	}

	txIndex, err := txIndexEnabled(client)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrFailedToDetectTxIndex, err)
	}

	currency, err := CurrencyFromChain(info.Chain)
	if err != nil {
		return nil, err
	}

	isNewWallet, err := loadOrCreateWallet(client)
	if err != nil {
		return nil, err
	}

	if isNewWallet {
		log.WithFields(log.Fields{
			"wallet": walletName,
		}).Info("Created new wallet")
	} else {
		log.WithFields(log.Fields{
			"wallet": walletName,
		}).Info("Loaded existing wallet")
	}

	b := &Bus{
		connChan:    pool,
		Pruned:      info.Pruned,
		Chain:       info.Chain,
		BlockFilter: blockFilter,
		TxIndex:     txIndex,
		Currency:    currency,
		Cache:       nil, // Disabled by default
		Status:      Initializing,
	}

	return b, nil
}

// Close performs cleanup operations on the Bus, notably shutting down the
// rpcclient connections.
func (b *Bus) Close() {
	if err := b.unloadWallet(); err != nil {
		log.WithFields(log.Fields{
			"wallet": walletName,
			"error":  err,
		}).Warn("Unable to unload wallet")
	}

	log.WithFields(log.Fields{
		"wallet": walletName,
	}).Info("Unloaded wallet successfully")

	close(b.connChan)
}

func (b *Bus) getClient() *rpcclient.Client {
	return <-b.connChan
}

func (b *Bus) recycleClient(client *rpcclient.Client) {
	b.connChan <- client
}

// Currency represents the currency type (btc) and the network params
// (Mainnet, testnet3, regtest, etc) in libcore parlance.
type Currency = string

const (
	Testnet Currency = "btc_testnet"
	Mainnet Currency = "btc"
)

// currencyFromChain is an adapter function to convert a chain (network) value
// to a Currency type that's understood by libcore.
func CurrencyFromChain(chain string) (Currency, error) {
	switch chain {
	case "regtest", "test":
		return Testnet, nil
	case "main":
		return Mainnet, nil
	default:
		return "", ErrUnrecognizedChain
	}
}

// loadOrCreateWallet attempts to load the default SatStack wallet, and if not
// found, creates the same.
//
// This method also detects if wallet features have been disabled in the
// Bitcoin node, and returns an error in such a case. This is typically the
// case when the option disablewallet=1 is specified in bitcoin.conf.
//
// The function returns a bool to indicate whether the wallet was created
// (true) or loaded (false). The value is meaningless if an error is returned.
//
// In case a new wallet is created, it'll be in loaded state by default.
func loadOrCreateWallet(client *rpcclient.Client) (bool, error) {
	// Try to load wallet first.
	_, err := client.LoadWallet(walletName)
	if err == nil {
		return false, nil
	}

	// Convert native error to btcjson.RPCError
	rpcErr := err.(*btcjson.RPCError)

	// Check if wallet RPC is disabled.
	if rpcErr.Code == btcjson.ErrRPCMethodNotFound.Code {
		return false, ErrWalletDisabled
	}

	if rpcErr.Code == btcjson.ErrRPCWallet && strings.Contains(rpcErr.Message, errDuplicateWalletLoadMsg) {
		// wallet already loaded. Ignore the error and return.
		return false, nil
	}

	// Wallet to load could not be found - create it.
	if rpcErr.Code == btcjson.ErrRPCWalletNotFound {
		if _, err := client.CreateWallet(
			walletName,
			rpcclient.WithCreateWalletDisablePrivateKeys(),
		); err != nil {
			return false, fmt.Errorf("%s: %w", ErrCreateWallet, err)
		}

		return true, nil
	}

	return false, fmt.Errorf("%s: %w", ErrLoadWallet, rpcErr)
}

// txIndexEnabled can be used to detect if the bitcoind server being connected
// has a transaction index (enabled by option txindex=1).
//
// It works by fetching the first (coinbase) transaction from the block at
// height 1. The genesis coinbase is normally not part of the transaction
// index, therefore we use block at height 1 instead of 0.
//
// If an irrecoverable error is encountered, it returns an error. In such
// cases, the caller may stop program execution.
func txIndexEnabled(client *rpcclient.Client) (bool, error) {
	blockHash, err := client.GetBlockHash(1)
	if err != nil {
		return false, ErrFailedToGetBlock
	}

	block, err := client.GetBlockVerbose(blockHash)
	if err != nil {
		return false, ErrFailedToGetBlock
	}

	if len(block.Tx) == 0 {
		return false, fmt.Errorf("no transactions in block at height 1")
	}

	// Coinbase transaction in block at height 1
	tx, err := utils.ParseChainHash(block.Tx[0])
	if err != nil {
		return false, fmt.Errorf(
			"%s (%s): %w", ErrMalformedChainHash, block.Tx[0], err)
	}

	if _, err := client.GetRawTransaction(tx); err != nil {
		return false, nil
	}

	return true, nil
}

// blockFilterEnabled can be used to detect if the bitcoind server being
// connected has a BIP-0157 Compact Block Filter index enabled (enabled by
// option blockfilterindex=1 in bitcoin.conf).
//
// Compact block filters use Golomb-Rice coding for compression and can give a
// probabilistic answer to the question "does the block contain x?". There are
// no false negatives.
//
// If this function returns true, blockchain rescans will be significantly
// faster, as bitcoind can avoid iterating through every transaction in every
// block.
func blockFilterEnabled(client *rpcclient.Client, hash string) (bool, error) {
	chainHash, err := chainhash.NewHashFromStr(hash)
	if err != nil {
		return false, err
	}

	if _, err := client.GetBlockFilter(*chainHash, nil); err != nil {
		return false, nil
	}

	return true, nil
}

func (b *Bus) WaitForNodeSync() error {
	client := b.getClient()
	defer b.recycleClient(client)

	b.Status = Syncing
	for {
		info, err := client.GetBlockChainInfo()
		if err != nil {
			return err
		}

		log.WithFields(log.Fields{
			"progress": fmt.Sprintf("%.2f%%", info.VerificationProgress*100),
		}).Info("Sychronizing node")

		if info.Blocks == info.Headers {
			log.WithFields(log.Fields{
				"blockHeight": info.Blocks,
				"blockHash":   info.BestBlockHash,
			}).Info("Sychronization complete")
			return nil
		}

		time.Sleep(10 * time.Second)
	}
}

// ImportAccounts will import the descriptors corresponding to the accounts
// into the Bitcoin Core wallet. This is a blocking operation.
func (b *Bus) ImportAccounts(accounts []config.Account) error {
	b.Status = Scanning

	var allDescriptors []descriptor
	for _, account := range accounts {
		accountDescriptors, err := b.descriptors(account)
		if err != nil {
			return err // return bare error, since it already has a ctx
		}

		allDescriptors = append(allDescriptors, accountDescriptors...)
	}

	var descriptorsToImport []descriptor
	for _, descriptor := range allDescriptors {
		address, err := b.DeriveAddress(descriptor.Value, descriptor.Depth)
		if err != nil {
			return fmt.Errorf("%s (%s - #%d): %w",
				ErrDeriveAddress, descriptor.Value, descriptor.Depth, err)
		}

		addressInfo, err := b.GetAddressInfo(*address)
		if err != nil {
			return fmt.Errorf("%s (%s): %w", ErrAddressInfo, *address, err)
		}

		if !addressInfo.IsWatchOnly {
			descriptorsToImport = append(descriptorsToImport, descriptor)
		}
	}

	if len(descriptorsToImport) == 0 {
		log.Warn("No (new) addresses to import")
		return nil
	}

	return b.ImportDescriptors(descriptorsToImport)
}

// descriptors returns canonical descriptors from the account configuration.
func (b *Bus) descriptors(account config.Account) ([]descriptor, error) {
	var ret []descriptor

	var depth int
	switch account.Depth {
	case nil:
		depth = defaultAccountDepth
	default:
		depth = *account.Depth
	}

	var age uint32
	switch account.Birthday {
	case nil:
		age = uint32(config.BIP0039Genesis.Unix())
	default:
		age = uint32(account.Birthday.Unix())
	}

	rawDescs := []string{
		strings.Split(*account.External, "#")[0], // strip out the checksum
		strings.Split(*account.Internal, "#")[0], // strip out the checksum
	}

	for _, desc := range rawDescs {
		canonicalDesc, err := b.GetCanonicalDescriptor(desc)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", ErrInvalidDescriptor, err)
		}

		ret = append(ret, descriptor{
			Value: *canonicalDesc,
			Depth: depth,
			Age:   age,
		})
	}

	return ret, nil
}

// RunTheNumbers performs inflation checks against the connected full node.
//
// It does NOT perform any equality comparison between expected and actual
// supply.
func (b *Bus) RunTheNumbers() error {
	client := b.getClient()
	defer b.recycleClient(client)

	log.Info("Running inflation checks...")

	info, err := client.GetTxOutSetInfo()
	if err != nil {
		return err
	}

	const halvingBlocks = 210000

	var (
		subsidy float64 = 50
		supply  float64 = 0
	)

	i := int64(0)
	for ; i < info.Height/halvingBlocks; i++ {
		supply += halvingBlocks * subsidy
		subsidy /= 2
	}

	supply += subsidy * float64(info.Height-(halvingBlocks*i))

	supplyBTC, err := btcutil.NewAmount(supply)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"height":         info.Height,
		"expectedSupply": supplyBTC,
		"actualSupply":   info.TotalAmount,
	}).Info("RunTheNumbers successful")

	return nil
}

func (b *Bus) unloadWallet() error {
	client := b.getClient()
	defer b.recycleClient(client)

	return client.UnloadWallet(nil)
}
