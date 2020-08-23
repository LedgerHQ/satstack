package main

import (
	"encoding/json"
	"ledger-sats-stack/bus"
	"ledger-sats-stack/httpd/svc"
	"os"
	"path"

	"ledger-sats-stack/config"
	"ledger-sats-stack/httpd"
	"ledger-sats-stack/version"

	"github.com/mitchellh/go-homedir"
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

	configuration := loadConfig()
	if configuration == nil {
		log.Fatal("Cannot find config file")
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

func loadConfig() *config.Configuration {
	home, err := homedir.Dir()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Cannot obtain user home directory")
	}

	configLookupPaths := []string{
		// TODO: Add Ledger Live user data folder
		"lss.json",
		path.Join(home, "lss.json"),
	}

	for _, configPath := range configLookupPaths {
		configuration, err := loadConfigFromPath(configPath)
		if err == nil {
			configuration.Validate()

			log.WithFields(log.Fields{
				"path": configPath,
			}).Info("Loaded config file")
			return &configuration
		}
	}

	return nil
}

func loadConfigFromPath(configPath string) (config.Configuration, error) {
	configuration := config.Configuration{}

	file, err := os.Open(configPath)
	if err != nil {
		return configuration, err
	}

	defer func() {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}()

	decoder := json.NewDecoder(file)

	err = decoder.Decode(&configuration)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Cannot decode accounts config JSON")
	}

	return configuration, nil
}
