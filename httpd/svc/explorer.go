package svc

import (
	"strconv"
	"time"

	"github.com/onyb/sat-stack/bus"

	"github.com/btcsuite/btcd/btcjson"
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

func (s *Service) GetStatus() (*bus.ExplorerStatus, error) {
	var syncProgress *float64
	if s.Bus.Status == bus.Syncing {
		info, err := s.Bus.GetBlockChainInfo()
		if err != nil {
			return nil, err
		}

		syncProgress = btcjson.Float64(info.VerificationProgress * 100)
	}

	var scanProgress *float64
	if s.Bus.Status == bus.Scanning {
		info, err := s.Bus.GetWalletInfo()
		if err != nil {
			return nil, err
		}

		switch v := info.Scanning.Value.(type) {
		case btcjson.ScanProgress:
			scanProgress = btcjson.Float64(v.Progress * 100)
		}
	}

	return &bus.ExplorerStatus{
		TxIndex:      s.Bus.TxIndex,
		Pruned:       s.Bus.Pruned,
		Chain:        s.Bus.Chain,
		Currency:     s.Bus.Currency,
		Status:       s.Bus.Status,
		SyncProgress: syncProgress,
		ScanProgress: scanProgress,
	}, nil
}
