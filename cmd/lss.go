package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/ledgerhq/satstack/bus"
	"github.com/ledgerhq/satstack/config"
	"github.com/ledgerhq/satstack/fortunes"
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

	fortunes.Fortune()

	s.Bus.Worker(configuration)

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

	log.Info("Shutdown server: in progress")

	{
		// Scoped block to disconnect all connections, and stop all goroutines.
		// If not successful within 5s, drop a nuclear bomb and fail with a
		// FATAL error.

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		s.Bus.Close(ctx)
	}

	{
		// Scoped block to gracefully shutdown Gin-Gonic server within 10s.

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.WithField("error", err).Fatal("Failed to shutdown server")
		}

		log.Info("Shutdown server: done")
	}
}
