package bus

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/btcsuite/btcd/chaincfg"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/ledgerhq/satstack/config"
	"github.com/ledgerhq/satstack/utils"
	"github.com/ledgerhq/satstack/version"
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
	minSupportedBitcoindVersion = 220000

	// walletName indicates the name of the wallet created by SatStack in
	// bitcoind's wallet.
	walletName = "satstack"

	errDuplicateWalletLoadMsg    = "Duplicate -wallet filename specified."
	errWalletAlreadyLoadedMsgOld = "Wallet file verification failed. Refusing to load database. Data file"
	// Cores Responds changes so adding the new one but keeping the old for backwards compatibility
	errWalletAlreadyLoadedMsgNew = "Wallet file verification failed. SQLiteDatabase: Unable to obtain an exclusive lock on the database"
)

var (
	// A new wallet needs to import the descriptors therefore
	// we need this information when starting the import worker
	isNewWallet bool
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

	// Thread-safe Bus cache, to query results typically by hash
	Cache *cache.Cache

	// Config to use for creating new connections on-demand.
	connCfg *rpcclient.ConnConfig

	// Primary RPC client for JSON-RPC requests. This does NOT allow batch
	// requests.
	mainClient *rpcclient.Client

	// Secondary RPC client for JSON-RPC requests. Use when mainClient is busy.
	secondaryClient *rpcclient.Client

	// RPC client reserved for performing RPC-based cleanups.
	janitorClient *rpcclient.Client

	// btcd network params
	Params *chaincfg.Params

	// IsPendingScan is a boolean field to indicate if satstack is currently
	// waiting for descriptors to be scanned or other initial operations like "running the numbers"
	// before the bridge can operate correctly
	//
	// This value can be exported for use by other packages to avoid making
	// explorer requests before satstack is able to serve them.
	IsPendingScan bool
}

type descriptor struct {
	Value string
	Depth int
	Age   uint32
}

// New initializes a Bus struct that embeds a btcd RPC client.
func New(host string, user string, pass string, proxy string, noTLS bool, unloadWallet bool) (*Bus, error) {
	log.Info("Warming up...")

	// Prepare the connection config to initialize the rpcclient.Client
	// pool with.
	connCfg := &rpcclient.ConnConfig{
		Host:         fmt.Sprintf("%s/wallet/%s", host, walletName),
		User:         user,
		Pass:         pass,
		Proxy:        proxy,
		HTTPPostMode: true,
		DisableTLS:   noTLS,
	}

	// Initialize RPC clients.
	mainClient, err := rpcclient.New(connCfg, nil)
	if err != nil {
		return nil, err // error ctx not required
	}

	secondaryClient, err := rpcclient.New(connCfg, nil)
	if err != nil {
		return nil, err // error ctx not required
	}

	janitorClient, err := rpcclient.New(connCfg, nil)
	if err != nil {
		return nil, err // error ctx not required
	}

	info, err := mainClient.GetBlockChainInfo()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrBitcoindUnreachable, err)
	}

	networkInfo, err := mainClient.GetNetworkInfo()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrBitcoindUnreachable, err)
	}

	if v := networkInfo.Version; v < minSupportedBitcoindVersion {
		return nil, fmt.Errorf("%s: %d", ErrUnsupportedBitcoindVersion, v)
	}

	blockFilter, err := blockFilterEnabled(mainClient, info.BestBlockHash)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrFailedToDetectBlockFilter, err)
	}

	txIndex, err := txIndexEnabled(mainClient)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrFailedToDetectTxIndex, err)
	}

	currency, err := CurrencyFromChain(info.Chain)
	if err != nil {
		return nil, err
	}

	if unloadWallet {
		if err = mainClient.UnloadWallet(nil); err != nil {
			return nil, err
		}

		log.Info("Unload wallet: done")
		os.Exit(1)
	}

	isNewWallet, err = loadOrCreateWallet(mainClient)
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

	params, err := ChainParams(info.Chain)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain params: %w", err)
	}

	b := &Bus{
		connCfg:         connCfg,
		mainClient:      mainClient,
		secondaryClient: secondaryClient,
		janitorClient:   janitorClient,
		Pruned:          info.Pruned,
		Chain:           info.Chain,
		BlockFilter:     blockFilter,
		TxIndex:         txIndex,
		Currency:        currency,
		Cache:           nil, // Disabled by default
		Params:          params,
		IsPendingScan:   true,
	}

	return b, nil
}

// Close performs cleanup operations on the Bus, notably shutting down the
// rpcclient.Client connections.
//
// The cleanup must be performed within a timeout set by the passed context,
// to prevent hanging on connections indefinitely held by bitcoind.
func (b *Bus) Close(ctx context.Context) {
	done := make(chan bool)

	go func() {
		b.mainClient.Shutdown()
		b.secondaryClient.Shutdown()

		// Only unload wallet if we are not in a pending scan
		// otherwise the nuclear timeout corrupts the wallet state
		if !b.IsPendingScan {
			b.UnloadWallet()
		}
		done <- true
	}()

	select {
	case <-ctx.Done():
		// Chernobyl nuclear disaster.

		b.janitorClient.Shutdown()
		log.WithField("error", ctx.Err()).Fatal("Shutdown server: force")
	case <-done:
		// The control rods have been lowered into the nuclear core, and the
		// chain reaction has gracefully stopped.
	}

}

func (b *Bus) ClientFactory() (*rpcclient.Client, error) {
	return rpcclient.New(b.connCfg, nil)
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

// ChainParams returns the *chaincfg.Params instance corresponding to the
// network that the underlying node is connected to.
//
// This value is useful for several operations in btcd, and can be accessed
// via the Bus struct.
func ChainParams(chain string) (*chaincfg.Params, error) {
	switch chain {
	case "regtest":
		return &chaincfg.RegressionNetParams, nil
	case "test":
		return &chaincfg.TestNet3Params, nil
	case "main":
		return &chaincfg.MainNetParams, nil
	default:
		return nil, ErrUnrecognizedChain
	}
}

type CreateWalletResult struct {
	Name    string `json:"name"`
	Warning string `json:"warning"`
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

	if rpcErr.Code == btcjson.ErrRPCWallet && strings.Contains(rpcErr.Message, errWalletAlreadyLoadedMsgOld) {
		// wallet already loaded. Ignore the error and return.
		return false, nil
	}

	if rpcErr.Code == btcjson.ErrRPCWallet && strings.Contains(rpcErr.Message, errWalletAlreadyLoadedMsgNew) {
		// wallet already loaded. Ignore the error and return.
		return false, nil
	}

	// Wallet to load could not be found - create it.
	if rpcErr.Code == btcjson.ErrRPCWalletNotFound {

		// see https://developer.bitcoin.org/reference/rpc/createwallet.html for specs and https://github.com/btcsuite/btcd/blob/3e2d8464f12b2e534e9764b0e4d4a48217c157e0/rpcclient/chain.go#L58 for example
		walletNameJSON, err := json.Marshal(walletName)
		if err != nil {
			return false, fmt.Errorf("%s: %w", "rawCreateWalletError walletNameJSON", err)
		}

		disablePrivateKeysJSON, err := json.Marshal(btcjson.Bool(true))
		if err != nil {
			return false, fmt.Errorf("%s: %w", "rawCreateWalletError disablePrivateKeysJSON", err)
		}

		blankJSON, err := json.Marshal(btcjson.Bool(true))
		if err != nil {
			return false, fmt.Errorf("%s: %w", "rawCreateWalletError blankJSON", err)
		}

		passphraseJSON, err := json.Marshal("")
		if err != nil {
			return false, fmt.Errorf("%s: %w", "rawCreateWalletError passphraseJSON", err)
		}

		avoidReuseJSON, err := json.Marshal(btcjson.Bool(false))
		if err != nil {
			return false, fmt.Errorf("%s: %w", "rawCreateWalletError avoidReuseJSON", err)
		}

		descriptorsJSON, err := json.Marshal(btcjson.Bool(true))
		if err != nil {
			return false, fmt.Errorf("%s: %w", "rawCreateWalletError descriptorsJSON", err)
		}

		loadOnStartupJSON, err := json.Marshal(btcjson.Bool(true))
		if err != nil {
			return false, fmt.Errorf("%s: %w", "rawCreateWalletError loadOnStartupJSON", err)
		}

		method := "createwallet"

		result, err := client.RawRequest(method, []json.RawMessage{
			walletNameJSON, disablePrivateKeysJSON, blankJSON, passphraseJSON, avoidReuseJSON, descriptorsJSON, loadOnStartupJSON,
		})

		if err != nil {
			return false, fmt.Errorf("%s: %w", "rawCreateWalletError err ", err)
		}

		var createWalletResult CreateWalletResult
		umerr := json.Unmarshal(result, &createWalletResult)

		if umerr != nil {
			return false, fmt.Errorf("%s: %w", "rawCreateWalletError umerr ", umerr)
		}

		log.Info(createWalletResult.Name + ` || ` + createWalletResult.Warning)

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

func (b *Bus) UnloadWallet() {
	if err := b.janitorClient.UnloadWallet(nil); err != nil {
		log.WithFields(log.Fields{
			"wallet": walletName,
			"error":  err,
		}).Warn("Unable to unload wallet")
		return
	}

	log.WithFields(log.Fields{
		"wallet": walletName,
	}).Info("Unloaded wallet successfully")

	b.janitorClient.Shutdown()
}

func (b *Bus) DumpLatestRescanTime() error {

	currentHeight, err := b.GetBlockCount()

	if err != nil {
		log.WithFields(log.Fields{
			"prefix": "worker",
		}).Error("Error fetching blockheight: %s", err)
		return err

	}
	data := &config.ConfigurationRescan{
		TimeStamp:       strconv.Itoa(int(time.Now().Unix())),
		LastSyncTime:    time.Now().Format(time.ANSIC),
		LastBlock:       currentHeight,
		SatstackVersion: version.Version,
	}
	err = config.WriteRescanConf(data)
	if err != nil {
		log.WithFields(log.Fields{
			"prefix": "worker",
		}).Errorf("Error saving last timestamp to file: %s", err)
		return err
	}

	return nil

}
