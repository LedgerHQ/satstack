package bus

// Status indicates the state of LSS with regards to the readiness of the
// connected Bitcoin Core node.
type Status string

const (
	// NodeDisconnected is a Status to indicate that the bitcoind instance is
	// unreachable. This is typically returned in the response of the status
	// endpoint.
	NodeDisconnected Status = "node-disconnected"

	// Ready is a Status to indicate that LSS is ready to accept explorer API
	// requests from Ledger Live.
	Ready Status = "ready"

	// Syncing is a Status to indicate that the Bitcoin Core node is currently
	// downloading and validating blocks.
	Syncing Status = "syncing"

	// Scanning is a Status to indicate that the Bitcoin Core node is currently
	// importing account descriptors into its wallet.
	Scanning Status = "scanning"
)

// ExplorerStatus represents the structure of payload returned by GetStatus
// service method.
type ExplorerStatus struct {
	TxIndex      bool     `json:"txindex"`
	BlockFilter  bool     `json:"block_filter"`
	Pruned       bool     `json:"pruned"`
	Chain        string   `json:"chain"`
	Currency     Currency `json:"currency"`
	Status       Status   `json:"status"`
	SyncProgress *float64 `json:"sync_progress,omitempty"`
	ScanProgress *float64 `json:"scan_progress,omitempty"`
}
