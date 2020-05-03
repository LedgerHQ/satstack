package transport

import (
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcjson"
	"ledger-sats-stack/pkg/utils"
	"strconv"
)

func (w Wire) GetHealth() error {
	_, err := w.GetBlockChainInfo()
	if err != nil {
		return err
	}

	// TODO: Check contents of GetBlockChainInfo response

	return nil
}

func (w Wire) GetSmartFeeEstimates(targets []int64, mode string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for _, target := range targets {

		fee, err := w.EstimateSmartFee(target, getMode(mode))
		if err != nil {
			return nil, err
		}
		if len(fee.Errors) > 0 {
			return nil, errors.New(fmt.Sprint(fee.Errors))
		}
		result[strconv.FormatInt(target, 10)] = utils.ParseSatoshi(*fee.FeeRate)
	}
	return result, nil
}

func getMode(s string) *btcjson.EstimateSmartFeeMode {
	switch s {
	case "UNSET":
		return &btcjson.EstimateModeUnset
	case "ECONOMICAL":
		return &btcjson.EstimateModeEconomical
	case "CONSERVATIVE":
		return &btcjson.EstimateModeConservative
	default:
		panic(s)
	}
}
