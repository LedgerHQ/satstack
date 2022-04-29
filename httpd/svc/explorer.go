package svc

import (
	"fmt"
	"strconv"
	"time"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/ledgerhq/satstack/bus"
	"github.com/ledgerhq/satstack/version"
	log "github.com/sirupsen/logrus"
)

func (s *Service) GetHealth() error {
	_, err := s.Bus.GetBlockChainInfo()
	if err != nil {
		return err
	}

	// TODO: Check contents of GetBlockChainInfo response

	return nil
}

func (s *Service) GetFees(targets []int64, mode string) map[string]interface{} {
	result := make(map[string]interface{})
	for _, target := range targets {
		fee := s.Bus.EstimateSmartFee(target, mode)
		result[strconv.FormatInt(target, 10)] = fee
	}

	result["last_updated"] = int32(time.Now().Unix())
	return result
}

func (s *Service) GetStatus() *bus.ExplorerStatus {
	// Prepare base bus.ExplorerStatus instance.
	status := bus.ExplorerStatus{
		Version:  version.Version,
		TxIndex:  s.Bus.TxIndex,
		Pruned:   s.Bus.Pruned,
		Chain:    s.Bus.Chain,
		Currency: s.Bus.Currency,
	}

	// Case 1: satstack is running the numbers.
	if s.Bus.IsPendingScan {
		status.Status = bus.PendingScan
		return &status
	}

	// Case 2: Unable to initialize rpcclient.Client.
	client, err := s.Bus.ClientFactory()
	if err != nil {
		log.WithField(
			"err", fmt.Errorf("%s: %w", bus.ErrBitcoindUnreachable, err),
		).Error("Failed to query status")
		status.Status = bus.NodeDisconnected
		return &status
	}

	defer client.Shutdown()

	// Case 3: bitcoind is unreachable - chain RPC failed.
	blockChainInfo, err := client.GetBlockChainInfo()
	if err != nil {
		log.WithField(
			"err", fmt.Errorf("%s: %w", bus.ErrBitcoindUnreachable, err),
		).Error("Failed to query status")

		status.Status = bus.NodeDisconnected
		return &status
	}

	// Case 4: bitcoind is currently catching up on new blocks.
	if blockChainInfo.Blocks != blockChainInfo.Headers {
		status.Status = bus.Syncing
		status.SyncProgress = btcjson.Float64(
			blockChainInfo.VerificationProgress * 100)
		return &status
	}

	// Case 5: bitcoind is currently importing descriptors
	walletInfo, err := client.GetWalletInfo()
	if err != nil {
		log.WithField(
			"err", fmt.Errorf("%s: %w", bus.ErrBitcoindUnreachable, err),
		).Error("Failed to query status")

		status.Status = bus.NodeDisconnected
		return &status
	}

	switch v := walletInfo.Scanning.Value.(type) {
	case btcjson.ScanProgress:
		status.Status = bus.Scanning
		status.ScanProgress = btcjson.Float64(v.Progress * 100)
		return &status
	}

	// Case 6: bitcoind is ready to be used with satstack.
	status.Status = bus.Ready
	return &status
}

func (s *Service) GetNetwork() (network *bus.Network) {
	client, err := s.Bus.ClientFactory()
	if err != nil {
		log.WithField("err", fmt.Errorf("%s: %w", bus.ErrBitcoindUnreachable, err)).
			Error("Failed to query status")

		network = new(bus.Network)
		return network
	}

	var networkInfo *btcjson.GetNetworkInfoResult
	if networkInfo, err = client.GetNetworkInfo(); err != nil {
		log.WithField("err", fmt.Errorf("%s: %w", bus.ErrBitcoindUnreachable, err)).
			Error("Failed to query status")

		network = new(bus.Network)
		return network
	}

	network = &bus.Network{
		RelayFee:       networkInfo.RelayFee,
		IncrementalFee: networkInfo.IncrementalFee,
		Version:        networkInfo.Version,
		Subversion:     network.Subversion,
	}
	return network
}
