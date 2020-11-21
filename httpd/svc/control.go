package svc

import (
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
