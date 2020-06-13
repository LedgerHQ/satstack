package svc

import (
	"fmt"
	"ledger-sats-stack/types"
	"ledger-sats-stack/utils"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// GetBlock is a service method to get a Block by a string reference
func (s *Service) GetBlock(ref string) (*types.Block, error) {
	rawBlockHash, err := s.getBlockHashByReference(ref)
	if err != nil {
		return nil, err
	}

	block, err := s.Bus.GetBlock(rawBlockHash)
	if err != nil {
		return nil, err
	}

	return block, nil
}

func (s *Service) getBlockHashByReference(ref string) (*chainhash.Hash, error) {
	switch {
	case ref == "current":
		return s.Bus.GetBestBlockHash()

	case strings.HasPrefix(ref, "0x"), len(ref) == 64:
		// 256-bit hex string with or without 0x prefix
		return utils.ParseChainHash(ref)
	default:
		{
			// Either an int64 block height, or garbage input
			blockHeight, err := strconv.ParseInt(ref, 10, 64)

			switch err {
			case nil:
				return s.Bus.GetBlockHash(blockHeight)

			default:
				return nil, fmt.Errorf("invalid block '%s'", ref)
			}
		}

	}
}
