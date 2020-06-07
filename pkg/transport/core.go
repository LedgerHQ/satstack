package transport

import (
	"errors"
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

	rawDescriptors := getAccountDescriptors(account)

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

	var requests []btcjson.ImportMultiRequest
	for _, descriptor := range allDescriptors {
		requests = append(requests, btcjson.ImportMultiRequest{
			Descriptor: btcjson.String(descriptor),
			Range:      &btcjson.DescriptorRange{Value: []int{0, depth}},
			Timestamp:  btcjson.Timestamp{Value: 0},
			WatchOnly:  btcjson.Bool(true),
			KeyPool:    btcjson.Bool(false),
			Internal:   btcjson.Bool(false),
		})
	}

	if requests == nil || len(requests) == 0 {
		err := errors.New("nothing to import")
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Import descriptors")
		return err
	}

	results, err := x.ImportMulti(requests, &btcjson.ImportMultiOptions{Rescan: true})
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Failed to import descriptors")
		return err
	}

	hasErrors := false

	for idx, result := range results {
		if result.Error != nil {
			log.WithFields(log.Fields{
				"descriptor": *requests[idx].Descriptor,
				"range":      requests[idx].Range.Value,
				"error":      result.Error,
			}).Error("Import output descriptor failed")
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
			}).Info("Import output descriptor")
		}
	}

	if hasErrors {
		return errors.New("failed to import descriptor")
	}

	return nil
}

func getAccountDescriptors(account config.Account) []string {
	var script string
	switch *account.DerivationMode { // cannot panic due to config validation
	case "standard":
		script = "pkh"
	case "segwit", "native_segwit":
		script = "wpkh"
	}

	var derivationPath string
	switch account.DerivationPath {
	case nil:
		// cannot panic due to config validation
		derivationPath = defaultAccountDerivationPaths[*account.DerivationMode]
	default:
		derivationPath = *account.DerivationPath
	}

	return []string{
		fmt.Sprintf("sh(%s([%s/%s/%d']%s/0/*))",
			script, masterKeyFingerprint, derivationPath, *account.Index, *account.XPub),
		fmt.Sprintf("sh(%s([%s/%s/%d']%s/1/*))",
			script, masterKeyFingerprint, derivationPath, *account.Index, *account.XPub),
	}
}
