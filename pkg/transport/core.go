package transport

import (
	"fmt"

	"ledger-sats-stack/pkg/config"

	"github.com/btcsuite/btcd/btcjson"
	log "github.com/sirupsen/logrus"
)

const masterKeyFingerprint = "d34db33f"
const defaultAccountDepth = 1000

var defaultAccountDerivationPaths = map[string]string{
	"standard":      "44'/60'",
	"segwit":        "49'/1'",
	"native_segwit": "84'/1'",
}

// GetCanonicalDescriptor returns the descriptor in canonical form, along with
// its computed checksum.
func (x XRPC) GetCanonicalDescriptor(descriptor string) (*string, error) {

	info, err := x.GetDescriptorInfo(descriptor)
	if err != nil {
		return nil, err
	}
	return &info.Descriptor, nil
}

func (x XRPC) getAccountDescriptors(account config.Account) ([]string, error) {
	var ret []string

	rawDescriptors := getRawAccountDescriptors(account)

	for _, desc := range rawDescriptors {
		canonicalDescriptor, err := x.GetCanonicalDescriptor(desc)
		if err != nil {
			return nil, err
		}

		ret = append(ret, *canonicalDescriptor)
	}

	return ret, nil
}

func (x XRPC) ImportAccounts(config config.Configuration) error {
	var depth int
	switch config.Depth {
	case nil:
		depth = defaultAccountDepth
	default:
		depth = *config.Depth
	}

	var allDescriptors []string
	for _, account := range config.Accounts {
		accountDescriptors, err := x.getAccountDescriptors(account)
		if err != nil {
			return err
		}
		allDescriptors = append(allDescriptors, accountDescriptors...)
	}

	var descriptorsToImport []string
	for _, descriptor := range allDescriptors {
		log.WithFields(log.Fields{
			"descriptor": descriptor,
			"depth":      depth,
		}).Info("Generate ranged descriptor")

		address := x.deriveFromDescriptor(descriptor, depth)
		addressInfo, err := x.GetAddressInfo(address)
		if err != nil {
			log.WithFields(log.Fields{
				"address": address,
				"error":   err,
			}).Fatal("Failed to get address info")
		}
		if !addressInfo.IsWatchOnly {
			descriptorsToImport = append(descriptorsToImport, descriptor)
		}
	}

	if len(descriptorsToImport) == 0 {
		log.Warn("No (new) addresses to import")
		return nil
	}

	var requests []btcjson.ImportMultiRequest
	for _, descriptor := range descriptorsToImport {
		requests = append(requests, btcjson.ImportMultiRequest{
			Descriptor: btcjson.String(descriptor),
			Range:      &btcjson.DescriptorRange{Value: []int{0, depth}},
			Timestamp:  btcjson.Timestamp{Value: 0}, // TODO: Use birthday here
			WatchOnly:  btcjson.Bool(true),
			KeyPool:    btcjson.Bool(false),
			Internal:   btcjson.Bool(false),
		})
	}

	log.WithFields(log.Fields{
		"rescan": true,
		"N":      len(requests),
	}).Info("Importing descriptors")

	results, err := x.ImportMulti(requests, &btcjson.ImportMultiOptions{Rescan: true})
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Failed to import descriptors")
	}

	hasErrors := false

	for idx, result := range results {
		if result.Error != nil {
			log.WithFields(log.Fields{
				"descriptor": *requests[idx].Descriptor,
				"range":      requests[idx].Range.Value,
				"error":      result.Error,
			}).Error("Failed to import descriptors")
			hasErrors = true
		}

		if result.Warnings != nil {
			log.WithFields(log.Fields{
				"descriptor": *requests[idx].Descriptor,
				"range":      requests[idx].Range.Value,
				"warnings":   result.Warnings,
			}).Warn("Import output descriptor")
		}

		if result.Success {
			log.WithFields(log.Fields{
				"descriptor": *requests[idx].Descriptor,
				"range":      requests[idx].Range.Value,
			}).Info("Import descriptor successful")
		}
	}

	if hasErrors {
		log.Fatal("Failed to import descriptors")
	}

	return nil
}

func getRawAccountDescriptors(account config.Account) []string {
	var scriptFragment string
	switch *account.DerivationMode { // cannot panic due to config validation
	case "standard":
		scriptFragment = "pkh(%s)"
	case "segwit":
		scriptFragment = "sh(wpkh(%s))"
	case "native_segwit":
		scriptFragment = "wpkh(%s)"
	}

	var derivationPath string
	switch account.DerivationPath {
	case nil:
		// cannot panic due to config validation
		derivationPath = defaultAccountDerivationPaths[*account.DerivationMode]
	default:
		derivationPath = *account.DerivationPath
	}

	accountDerivationPath := fmt.Sprintf("%s/%s/%d'",
		masterKeyFingerprint, derivationPath, *account.Index)

	return []string{
		fmt.Sprintf(scriptFragment, // external chain (receive address descriptor)
			fmt.Sprintf("[%s]%s/0/*", accountDerivationPath, *account.XPub)),

		fmt.Sprintf(scriptFragment, // internal chain (change address descriptor)
			fmt.Sprintf("[%s]%s/1/*", accountDerivationPath, *account.XPub)),
	}
}

func (x XRPC) deriveFromDescriptor(descriptor string, addressIndex int) string {
	addresses, err := x.DeriveAddresses(
		descriptor,

		// Since we're interested in only the address at addressIndex,
		// specifying the range as [begin, end] ensures that addresses
		// between index 0 and end-1 are not derived uselessly.
		&btcjson.DescriptorRange{Value: []int{addressIndex, addressIndex}},
	)
	if err != nil {
		log.WithFields(log.Fields{
			"error":      err,
			"descriptor": descriptor,
			"index":      addressIndex,
		}).Fatal("Failed to derive address")
	}

	return (*addresses)[0] // *addresses is always a single-element slice
}
