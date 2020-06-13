package svc

import (
	"fmt"
	"ledger-sats-stack/config"

	log "github.com/sirupsen/logrus"
)

const masterKeyFingerprint = "d34db33f"
const defaultAccountDepth = 1000

var defaultAccountDerivationPaths = map[string]string{
	"standard":      "44'/60'",
	"segwit":        "49'/1'",
	"native_segwit": "84'/1'",
}

func (s *Service) ImportAccounts(config config.Configuration) error {
	var depth int
	switch config.Depth {
	case nil:
		depth = defaultAccountDepth
	default:
		depth = *config.Depth
	}

	var allDescriptors []string
	for _, account := range config.Accounts {
		accountDescriptors, err := s.getAccountDescriptors(account)
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

		address, err := s.deriveFromDescriptor(descriptor, depth)
		if err != nil {
			log.WithFields(log.Fields{
				"error":      err,
				"descriptor": descriptor,
				"index":      depth,
			}).Error("Failed to derive address")
			return err
		}
		addressInfo, err := s.Bus.GetAddressInfo(*address)
		if err != nil {
			log.WithFields(log.Fields{
				"address": *address,
				"error":   err,
			}).Error("Failed to get address info")
			return err
		}
		if !addressInfo.IsWatchOnly {
			descriptorsToImport = append(descriptorsToImport, descriptor)
		}
	}

	if len(descriptorsToImport) == 0 {
		log.Warn("No (new) addresses to import")
		return nil
	}

	return s.Bus.ImportDescriptors(descriptorsToImport, depth)
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

func (s *Service) deriveFromDescriptor(descriptor string, addressIndex int) (*string, error) {
	address, err := s.Bus.DeriveAddress(descriptor, addressIndex)
	if err != nil {
		return nil, err
	}

	return address, nil
}

func (s *Service) getAccountDescriptors(account config.Account) ([]string, error) {
	var ret []string

	rawDescriptors := getRawAccountDescriptors(account)

	for _, desc := range rawDescriptors {
		canonicalDescriptor, err := s.Bus.GetCanonicalDescriptor(desc)
		if err != nil {
			return nil, err
		}

		ret = append(ret, *canonicalDescriptor)
	}

	return ret, nil
}
