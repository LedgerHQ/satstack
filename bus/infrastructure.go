package bus

import (
	"fmt"
	"ledger-sats-stack/config"
	"ledger-sats-stack/types"
	"ledger-sats-stack/utils"
	"strings"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	log "github.com/sirupsen/logrus"
)

const defaultAccountDepth = 1000

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

// TxIndexEnabled can be used to detect if the bitcoind server being connected
// has a transaction index (enabled by option txindex=1).
//
// It works by fetching the first (coinbase) transaction from the block at
// height 1. The genesis coinbase is normally not part of the transaction
// index, therefore we use block at height 1 instead of 0.
//
// If an irrecoverable error is encountered, it returns an error. In such
// cases, the caller may stop program execution.
func TxIndexEnabled(client *rpcclient.Client) (bool, error) {
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

	var allDescriptors []types.Descriptor
	for _, account := range accounts {
		accountDescriptors, err := b.Descriptors(account)
		if err != nil {
			return err // return bare error, since it already has a ctx
		}

		allDescriptors = append(allDescriptors, accountDescriptors...)
	}

	var descriptorsToImport []types.Descriptor
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

// Descriptors returns canonical descriptors from the account configuration.
func (b *Bus) Descriptors(account config.Account) ([]types.Descriptor, error) {
	var ret []types.Descriptor

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

		ret = append(ret, types.Descriptor{
			Value: *canonicalDesc,
			Depth: depth,
			Age:   age,
		})
	}

	return ret, nil
}
