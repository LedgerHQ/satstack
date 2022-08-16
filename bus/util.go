package bus

import (
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/ledgerhq/satstack/utils"
	log "github.com/sirupsen/logrus"
)

const fallbackFee = btcutil.Amount(1)

func (b *Bus) EstimateSmartFee(target int64, mode string) btcutil.Amount {
	fee, err := b.mainClient.EstimateSmartFee(target, getMode(mode))

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
		}).Error("Failed estimatesmartfee Bridge")
		return fallbackFee
	}

	if len(fee.Errors) > 0 {
		log.WithFields(log.Fields{
			"error":  fee.Errors,
			"target": target,
			"mode":   mode,
		}).Error("Failed estimatesmartfee Bridge")
		return fallbackFee
	}

	return utils.ParseSatoshi(*fee.FeeRate)
}

func DeriveAddress(client *rpcclient.Client, descriptor string, index int) (*string, error) {
	addresses, err := client.DeriveAddresses(
		descriptor,

		// Since we're interested in only the address at addressIndex,
		// specifying the range as [begin, end] ensures that addresses
		// between index 0 and end-1 are not derived uselessly.
		&btcjson.DescriptorRange{Value: []int{index, index}},
	)
	if err != nil {
		log.WithFields(log.Fields{
			"error":      err,
			"descriptor": descriptor,
			"index":      index,
		}).Error("Failed to derive address")
		return nil, err
	}

	return &(*addresses)[0], nil // *addresses is always a single-element slice
}

// GetCanonicalDescriptor returns the descriptor in canonical form, along with
// its computed checksum.
func GetCanonicalDescriptor(client *rpcclient.Client, descriptor string) (*string, error) {
	info, err := client.GetDescriptorInfo(descriptor)
	if err != nil {
		return nil, err
	}
	return &info.Descriptor, nil
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
		return &btcjson.EstimateModeEconomical
	}
}
