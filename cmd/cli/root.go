package cli

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
	"github.com/spf13/cobra"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

func init() {
	rootCmd.PersistentFlags().String("port", "20000", "Port")
	rootCmd.PersistentFlags().Bool("unload-wallet", false, "whether SatStack should unload wallet")
	rootCmd.PersistentFlags().Bool("circulation-check", false, "performs inflation checks against the connected full node")
	rootCmd.PersistentFlags().Bool("force-importdescriptors", false, "this will force importing descriptors although the wallet does already exist "+
		"which will force the wallet to rescan from the brithday date")

}

var rootCmd = &cobra.Command{
	Use:   "lss",
	Short: "Bitcoin full node with Ledger Live.",
	Long:  `Ledger SatStack is a lightweight bridge to connect Ledger Live with your personal Bitcoin full node. It's designed to allow Ledger Live users use Bitcoin without compromising on privacy, or relying on Ledger's infrastructure.`,
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetString("port")
		unloadWallet, _ := cmd.Flags().GetBool("unload-wallet")
		circulationCheck, _ := cmd.Flags().GetBool("circulation-check")
		forceImportDesc, _ := cmd.Flags().GetBool("force-importdescriptors")

		s := startup(unloadWallet, circulationCheck, forceImportDesc)
		if s == nil {
			return
		}

		engine := httpd.GetRouter(s)

		srv := &http.Server{
			Addr:    ":" + port,
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

			// In case we are scanning the wallet, we have to abort the wallet
			// because unloading the wallet while scanning will result in a timeout
			// and a non recoverable state. This will be fixed by
			// https://github.com/bitcoin/bitcoin/pull/26618

			if s.Bus.IsPendingScan {

				err := s.Bus.AbortRescan()
				if err != nil {
					log.WithFields(log.Fields{
						"error": err,
					}).Error("Failed to abort rescan")
				}
			} else {
				err := s.Bus.DumpLatestRescanTime()
				if err != nil {
					log.WithFields(log.Fields{
						"prefix": "worker",
						"error":  err,
					}).Error("Failed to dump latest block into file")
				}
			}

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
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func startup(unloadWallet bool, circulationCheck bool, forceImportDesc bool) *svc.Service {

	if version.Build == "development" {
		log.SetLevel(log.DebugLevel)
	}

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
		configuration.TorProxy,
		configuration.NoTLS,
		unloadWallet,
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

	s.Bus.Worker(configuration, circulationCheck, forceImportDesc)

	return s
}
