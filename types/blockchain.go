package types

// BlockChainInfo models the data from the getblockchaininfo command.
//
// The fields are explicitly defined here because the `softforks` field is
// a map in the btcd library, but a slice in the Bitcoin Core RPC. This was
// fixed in btcd master, but the latest release (v0.22.1) still has the bug.
//
// See https://github.com/btcsuite/btcd/pull/1676
// See https://github.com/btcsuite/btcd/pull/1814
type BlockChainInfo struct {
	Chain                string               `json:"chain"`
	Blocks               int32                `json:"blocks"`
	Headers              int32                `json:"headers"`
	BestBlockHash        string               `json:"bestblockhash"`
	Difficulty           float64              `json:"difficulty"`
	MedianTime           int64                `json:"mediantime"`
	VerificationProgress float64              `json:"verificationprogress"`
	InitialBlockDownload bool                 `json:"initialblockdownload"`
	ChainWork            string               `json:"chainwork"`
	SizeOnDisk           int64                `json:"size_on_disk"`
	Pruned               bool                 `json:"pruned"`
	PruneHeight          int32                `json:"pruneheight,omitempty"`
	AutomaticPruning     bool                 `json:"automatic_pruning,omitempty"`
	PruneTargetSize      int64                `json:"prune_target_size,omitempty"`
	SoftForks            []*SoftFork          `json:"softforks"`
	Warnings             []string             `json:"warnings"`
}

// SoftFork describes a soft fork.
//
// See BlockChainInfo for why this is defined here.
type SoftFork struct {
	ID      string          `json:"id"`
	Version int32           `json:"version"`
	Reject  *SoftForkReject `json:"reject"`
}

// SoftForkReject describes a soft fork rejection.
//
// See BlockChainInfo for why this is defined here.
type SoftForkReject struct {
	Status bool `json:"status"`
}
