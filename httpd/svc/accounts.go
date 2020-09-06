package svc

import (
	"ledger-sats-stack/config"
	"ledger-sats-stack/types"
	"strings"

	log "github.com/sirupsen/logrus"
)

const defaultAccountDepth = 1000

func (s *Service) ImportAccounts(config config.Configuration) error {
	var allDescriptors []types.Descriptor
	for _, account := range config.Accounts {
		accountDescriptors, err := s.getAccountDescriptors(account)
		if err != nil {
			return err
		}

		allDescriptors = append(allDescriptors, accountDescriptors...)
	}

	var descriptorsToImport []types.Descriptor
	for _, descriptor := range allDescriptors {
		log.WithFields(log.Fields{
			"descriptor": descriptor.Value,
			"depth":      descriptor.Depth,
			"age":        descriptor.Age,
		}).Info("Generate ranged descriptor")

		address, err := s.deriveFromDescriptor(descriptor.Value, descriptor.Depth)
		if err != nil {
			log.WithFields(log.Fields{
				"error":      err,
				"descriptor": descriptor,
				"index":      descriptor.Depth,
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

	return s.Bus.ImportDescriptors(descriptorsToImport)
}

func getRawAccountDescriptors(account config.Account) []string {
	return []string{
		strings.Split(*account.External, "#")[0], // strip out the checksum
		strings.Split(*account.Internal, "#")[0], // strip out the checksum
	}
}

func (s *Service) deriveFromDescriptor(descriptor string, addressIndex int) (*string, error) {
	address, err := s.Bus.DeriveAddress(descriptor, addressIndex)
	if err != nil {
		return nil, err
	}

	return address, nil
}

func (s *Service) getAccountDescriptors(account config.Account) ([]types.Descriptor, error) {
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
		age = uint32(config.LedgerNanoSGenesis.Unix())
	default:
		age = uint32(account.Birthday.Unix())
	}

	rawDescriptors := getRawAccountDescriptors(account)

	for _, desc := range rawDescriptors {
		canonicalDescriptor, err := s.Bus.GetCanonicalDescriptor(desc)
		if err != nil {
			return nil, err
		}

		ret = append(ret, types.Descriptor{
			Value: *canonicalDescriptor,
			Depth: depth,
			Age:   age,
		})
	}

	return ret, nil
}
