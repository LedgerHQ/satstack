package svc

import (
	"strconv"
	"time"
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
