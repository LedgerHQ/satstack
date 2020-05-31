package transport

import (
	"errors"
	"fmt"

	"ledger-sats-stack/pkg/types"

	"github.com/btcsuite/btcd/btcjson"
	log "github.com/sirupsen/logrus"
)

const masterKeyFingerprint = "d34db33f"

const legacyDerivationPath = "44'/60'"
const segwitDerivationPath = "49'/1'"
const nativeSegwitDerivationPath = "84'/1'"

// GetCanonicalDescriptor returns the descriptor in canonical form, along with
// its computed checksum.
func (x XRPC) GetCanonicalDescriptor(descriptor string) (*string, error) {
	info, err := x.GetDescriptorInfo(descriptor)
	if err != nil {
		return nil, err
	}
	return &info.Descriptor, nil
}

func (x XRPC) getAccountDescriptors(account types.Account) ([]string, error) {
	var ret []string

	rawDescriptors := func() []string {
		switch account.Type {
		case "legacy":
			return []string{
				fmt.Sprintf("sh(pkh([%s/%s/%d']%s/0/*))",
					masterKeyFingerprint, legacyDerivationPath, account.Index, account.XPub),
				fmt.Sprintf("sh(pkh([%s/%s/%d']%s/1/*))",
					masterKeyFingerprint, legacyDerivationPath, account.Index, account.XPub),
			}

		case "segwit":
			return []string{
				fmt.Sprintf("sh(wpkh([%s/%s/%d']%s/0/*))",
					masterKeyFingerprint, segwitDerivationPath, account.Index, account.XPub),
				fmt.Sprintf("sh(wpkh([%s/%s/%d']%s/1/*))",
					masterKeyFingerprint, segwitDerivationPath, account.Index, account.XPub),
			}

		case "native_segwit":
			return []string{
				fmt.Sprintf("sh(wpkh([%s/%s/%d']%s/0/*))",
					masterKeyFingerprint, nativeSegwitDerivationPath, account.Index, account.XPub),
				fmt.Sprintf("sh(wpkh([%s/%s/%d']%s/1/*))",
					masterKeyFingerprint, nativeSegwitDerivationPath, account.Index, account.XPub),
			}

		default:
			return []string{}
		}
	}()

	for _, desc := range rawDescriptors {
		canonicalDescriptor, err := x.GetCanonicalDescriptor(desc)
		if err != nil {
			return nil, err
		}
		ret = append(ret, *canonicalDescriptor)
	}

	return ret, nil
}

func (x XRPC) ImportAccounts(accounts []types.Account) error {
	var allDescriptors []string
	for _, account := range accounts {
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
			Range:      &btcjson.DescriptorRange{Value: []int{0, 1000}},
			Timestamp:  btcjson.Timestamp{Value: 0},
			WatchOnly:  btcjson.Bool(true),
			KeyPool:    btcjson.Bool(false),
			Internal:   btcjson.Bool(true),
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
