package main

import (
	"github.com/ledgerhq/satstack/bus"
	"github.com/ledgerhq/satstack/config"
	"github.com/ledgerhq/satstack/httpd"
	"github.com/ledgerhq/satstack/httpd/svc"
	"github.com/ledgerhq/satstack/version"

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
	}).Infof("Ledger Sat Stack (lss) %s", version.Version)

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

	s := &svc.Service{
		Bus: b,
	}

	go func() {
		if err := b.WaitForNodeSync(); err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Fatal("Failed during node sync")
		}

		if err := b.ImportAccounts(configuration.Accounts); err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Fatal("Failed to import accounts")
		}

		b.Status = bus.Ready
	}()

	engine := httpd.GetRouter(s)

	_ = engine.Run(":20000")
}
