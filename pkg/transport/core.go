package transport

import (
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/btcjson"
	log "github.com/sirupsen/logrus"
)

// GetCanonicalDescriptor returns the descriptor in canonical form, along with
// its computed checksum.
func (w Wire) GetCanonicalDescriptor(descriptor string) (*string, error) {
	info, err := w.GetDescriptorInfo(descriptor)
	if err != nil {
		return nil, err
	}
	return &info.Descriptor, nil
}

func (w Wire) getDescriptorsFromXpub(xpub string) ([]string, error) {
	var ret []string
	for _, desc := range []string{
		fmt.Sprintf("pkh(%s/0/*)", xpub),
		fmt.Sprintf("pkh(%s/1/*)", xpub),
		fmt.Sprintf("wpkh(%s/0/*)", xpub),
		fmt.Sprintf("wpkh(%s/1/*)", xpub),
	} {
		canonicalDescriptor, err := w.GetCanonicalDescriptor(desc)
		if err != nil {
			return nil, err
		}
		ret = append(ret, *canonicalDescriptor)
	}
	return ret, nil
}

func (w Wire) ImportXpubs(xpubs []string) error {
	var allDescriptors []string
	for _, xpub := range xpubs {
		xpubDescriptors, err := w.getDescriptorsFromXpub(xpub)
		if err != nil {
			return err
		}
		allDescriptors = append(allDescriptors, xpubDescriptors...)
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

	results, err := w.ImportMulti(requests, &btcjson.ImportMultiOptions{Rescan: true})
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
