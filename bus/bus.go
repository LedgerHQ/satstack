package bus

import "github.com/btcsuite/btcd/rpcclient"

type Bus struct {
	client   *rpcclient.Client
	Chain    string
	Pruned   bool
	TxIndex  bool
	Currency string // Based on Chain value, for interoperability with libcore
}
