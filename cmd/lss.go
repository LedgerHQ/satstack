package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/ledgerhq/satstack/bus"
	"github.com/ledgerhq/satstack/config"
	"github.com/ledgerhq/satstack/httpd"
	"github.com/ledgerhq/satstack/httpd/svc"
	"github.com/ledgerhq/satstack/version"
	log "github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

func startup() *svc.Service {
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
	}).Infof("Ledger SatStack (lss) %s", version.Version)

	configuration, err := config.Load()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatal("Failed to load config")
		return nil
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
		return nil
	}

	log.WithFields(log.Fields{
		"chain":       b.Chain,
		"pruned":      b.Pruned,
		"txindex":     b.TxIndex,
		"blockFilter": b.BlockFilter,
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

		if err := b.RunTheNumbers(); err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Fatal("Inflation checks failed")
		}

		// Skip import of descriptors, if no account config found. SatStack
		// will run in zero-configuration mode.
		if configuration.Accounts == nil {
			return
		}

		if err := b.ImportAccounts(configuration.Accounts); err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Fatal("Failed to import accounts")
		}
	}()

	return s
}

func main() {
	s := startup()
	engine := httpd.GetRouter(s)

	srv := &http.Server{
		Addr:    ":20000",
		Handler: engine,
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithFields(log.Fields{
				"error": err,
			}).Fatal("Failed to listen and serve")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit

	fmt.Println()
	log.Info("Shutting down server...")

	s.Bus.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
}
