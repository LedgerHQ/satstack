package main

import (
	"ledger-sats-stack/bus"
	"ledger-sats-stack/config"
	"ledger-sats-stack/httpd"
	"ledger-sats-stack/httpd/svc"
	"ledger-sats-stack/version"

	log "github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

func main() {
	log.SetFormatter(&prefixed.TextFormatter{
		TimestampFormat:  "2006/01/02 - 15:04:05",
		FullTimestamp:    true,
		QuoteEmptyFields: true,
		SpacePadding:     45,
	})

	log.WithFields(log.Fields{
		"build":   version.Build,
		"commit":  version.GitCommit,
		"runtime": version.GoVersion,
		"arch":    version.OsArch,
	}).Infof("Ledger Sats Stack (lss) %s", version.Version)

	configuration, err := config.Load()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatal("Failed to load config")
		return
	}

	b, err := bus.New(
		*configuration.RPCURL,
		*configuration.RPCUser,
		*configuration.RPCPassword,
		configuration.NoTLS,
	)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Failed to initialize Bus")
		return
	}
	defer b.Close()

	log.WithFields(log.Fields{
		"chain":   b.Chain,
		"pruned":  b.Pruned,
		"txindex": b.TxIndex,
	}).Info("RPC connection established")

	bus.WaitForNodeSync(b)

	s := &svc.Service{
		Bus: b,
	}

	if err := s.ImportAccounts(*configuration); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Failed to import accounts")
	}

	engine := httpd.GetRouter(s)
	_ = engine.Run(":20000")
}
