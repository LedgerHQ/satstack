package bus

import "errors"

var (
	// ErrBitcoindUnreachable indicates that an RPC call to the bitcoind node
	// was not successful. Use this error during sanity checks.
	ErrBitcoindUnreachable = errors.New("bitcoind unreachable")

	// ErrWalletDisabled indicates that wallet features have been disabled on
	// the connected Bitcoin node. SatStack relies on wallet RPCs to function.
	ErrWalletDisabled = errors.New("bitcoind wallet is disabled")

	// ErrWalletRPCFailed indicates that a wallet RPC did not succeed.
	ErrWalletRPCFailed = errors.New("bitcoind wallet RPC failed")

	// ErrUnsupportedBitcoindVersion indicates that the connected bitcoind node
	// has a version that is not supported by SatStack.
	ErrUnsupportedBitcoindVersion = errors.New("unsupported bitcoind version")

	// ErrUnrecognizedChain indicates that the Chain returned by bitcoind in
	// its response to the getblockchaininfo RPC, is unrecognized by LSS.
	//
	// This usually means that the value doesn't correspond to a Currency or
	// network that libcore can understand.
	ErrUnrecognizedChain = errors.New("unrecognized chain")

	// ErrFailedToGetBlock indicates that an error was encountered while
	// trying to get a block.
	ErrFailedToGetBlock = errors.New("failed to get block")

	// ErrMalformedChainHash indicates that a chain hash (transaction or block)
	// could not be parsed.
	ErrMalformedChainHash = errors.New("malformed chain hash")

	// ErrFailedToDetectTxIndex indicates an irrecoverable error while trying
	// to detect presence of a transaction index. Normally, this error should
	// not be ignored silently.
	ErrFailedToDetectTxIndex = errors.New("failed to detect txindex")

	// ErrFailedToDetectBlockFilter indicates an irrecoverable error while trying
	// to detect presence of a compact block filter index. Normally, this error
	// should not be ignored silently.
	ErrFailedToDetectBlockFilter = errors.New("failed to detect block filter")

	// ErrInvalidDescriptor indicates that a malformed descriptor was
	// encountered.
	ErrInvalidDescriptor = errors.New("invalid descriptor")

	// ErrDeriveAddress indicates that an address could not be derived from a
	// descriptor.
	ErrDeriveAddress = errors.New("failed to derive address")

	// ErrAddressInfo indicates that an error was encountered while trying to
	// fetch address info.
	ErrAddressInfo = errors.New("failed to get address info")
)
