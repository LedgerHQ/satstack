package transport

import (
	"strconv"
	"time"

	"ledger-sats-stack/pkg/utils"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcutil"
	log "github.com/sirupsen/logrus"
)

const fallbackFee = btcutil.Amount(1)

func (w Wire) GetHealth() error {
	_, err := w.GetBlockChainInfo()
	if err != nil {
		return err
	}

	// TODO: Check contents of GetBlockChainInfo response

	return nil
}

func (w Wire) GetSmartFeeEstimates(targets []int64, mode string) map[string]interface{} {
	result := make(map[string]interface{})
	for _, target := range targets {
		fee := w.getFeeForTarget(target, mode)
		result[strconv.FormatInt(target, 10)] = fee
	}

	result["last_updated"] = int32(time.Now().Unix())
	return result
}

func (w Wire) getFeeForTarget(target int64, mode string) btcutil.Amount {
	fee, err := w.EstimateSmartFee(target, getMode(mode))

	// If failed to get smart fee estimate, fallback to fallbackFee.
	// Example: if the full-node is a regtest chain, there are normally
	// no transactions in the mempool to analyze for estimating fees.
	//
	// TODO: Use Minimum Relay Fee instead of btcutil.Amount(1)
	if err != nil {
		log.WithFields(log.Fields{
			"error":  err,
			"target": target,
			"mode":   mode,
		}).Error("Failed estimatesmartfee RPC")
		return fallbackFee
	}

	if len(fee.Errors) > 0 {
		log.WithFields(log.Fields{
			"error":  fee.Errors,
			"target": target,
			"mode":   mode,
		}).Error("Failed estimatesmartfee RPC")
		return fallbackFee
	}

	return utils.ParseSatoshi(*fee.FeeRate)
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
