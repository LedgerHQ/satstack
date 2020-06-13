package svc

import (
	"ledger-sats-stack/pkg/config"
	"ledger-sats-stack/pkg/types"
)

type TransactionsService interface {
	GetTransaction(hash string) (*types.Transaction, error)
	GetTransactionHex(hash string) (*string, error)
	SendTransaction(tx string) (*string, error)
}

type BlocksService interface {
	GetBlock(ref string) (*types.Block, error)
}

type AddressesService interface {
	GetAddresses(addresses []string) (types.Addresses, error)
}

type ExplorerService interface {
	GetHealth() error
	GetFees(targets []int64, mode string) map[string]interface{}
}

type CoreService interface {
	ImportAccounts(config config.Configuration) error
}

type ServiceInterface interface {
	BlocksService
	TransactionsService
	AddressesService
	ExplorerService
	CoreService
}
