package svc

import (
	"github.com/ledgerhq/satstack/bus"
	"github.com/ledgerhq/satstack/config"
	"github.com/ledgerhq/satstack/types"
)

type TransactionsService interface {
	GetTransaction(hash string, block *types.Block, bestBlockHeight int32) (*types.Transaction, error)
	GetTransactionHex(hash string) (string, error)
	SendTransaction(tx string) (string, error)
}

type BlocksService interface {
	GetBlock(ref string) (*types.Block, error)
}

type AddressesService interface {
	GetAddresses(addresses []string, blockHash *string) (types.Addresses, error)
}

type ExplorerService interface {
	GetFees(targets []int64, mode string) map[string]interface{}
	GetHealth() error
	GetNetwork() *bus.Network
	GetStatus() *bus.ExplorerStatus
}

type ControlService interface {
	HasDescriptor(descriptor string) (bool, error)
	ImportAccounts(accounts []config.Account)
}

type ServiceInterface interface {
	AddressesService
	BlocksService
	ControlService
	ExplorerService
	TransactionsService
}
