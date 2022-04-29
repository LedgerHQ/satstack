package bus

type Network struct {
	RelayFee       float64 `json:"relay_fee"`
	IncrementalFee float64 `json:"incremental_fee"`
	Version        int32   `json:"version"`
	Subversion     string  `json:"subversion"`
}
