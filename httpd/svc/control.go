package svc

import (
	"fmt"

	"github.com/ledgerhq/satstack/bus"
	"github.com/ledgerhq/satstack/config"
	log "github.com/sirupsen/logrus"
)

func (s *Service) ImportAccounts(accounts []config.Account) {
	go func() {
		if err := s.Bus.ImportAccounts(accounts); err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Failed to import accounts")
		}
	}()
}

func (s *Service) HasDescriptor(descriptor string) (bool, error) {
	client, err := s.Bus.ClientFactory()
	if err != nil {
		return false, err
	}

	defer client.Shutdown()

	canonicalDesc, err := bus.GetCanonicalDescriptor(client, descriptor)
	if err != nil {
		return false, fmt.Errorf("%s: %w", bus.ErrInvalidDescriptor, err)
	}

	address, err := bus.DeriveAddress(client, *canonicalDesc, 0)
	if err != nil {
		return false, fmt.Errorf("%s (%s - #%d): %w",
			bus.ErrDeriveAddress, *canonicalDesc, 0, err)
	}

	addressInfo, err := client.GetAddressInfo(*address)
	if err != nil {
		return false, fmt.Errorf("%s (%s): %w", bus.ErrAddressInfo, *address, err)
	}

	if !addressInfo.IsWatchOnly {
		return false, nil
	}

	return true, nil
}
