package bus

import (
	"encoding/json"

	"github.com/ledgerhq/satstack/types"
	"github.com/ledgerhq/satstack/utils"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

func (b *Bus) GetBestBlockHash() (*chainhash.Hash, error) {
	return b.mainClient.GetBestBlockHash()
}

func (b *Bus) GetBlockCount() (int64, error) {
	return b.mainClient.GetBlockCount()

}

func (b *Bus) GetBlockHash(height int64) (*chainhash.Hash, error) {
	return b.mainClient.GetBlockHash(height)
}

func (b *Bus) GetBlock(hash *chainhash.Hash) (*types.Block, error) {
	nativeBlock, err := b.mainClient.GetBlockVerbose(hash)
	if err != nil {
		return nil, err
	}

	transactions := make([]string, len(nativeBlock.Tx))
	for idx, transaction := range nativeBlock.Tx {
		transactions[idx] = transaction
	}

	block := types.Block{
		Hash:         nativeBlock.Hash,
		Height:       nativeBlock.Height,
		Time:         utils.ParseUnixTimestamp(nativeBlock.Time),
		Transactions: &transactions,
	}

	return &block, nil
}

func (b *Bus) GetBlockChainInfo() (*types.BlockChainInfo, error) {
	// The `softforks` field is a map in the btcd library, but a slice in
	// the Bitcoin Core RPC. This was fixed in btcd master, but the latest
	// release (v0.22.1) still has the bug.
	//
	// The `warnings` field is a string in the btcd library, but a slice
	// in the Bitcoin Core RPC.
	//
	// As a result, we have to bypass the btcd library and use a raw request
	// to get the blockchain info.
	//
	// See https://github.com/btcsuite/btcd/pull/1676
	// See https://github.com/btcsuite/btcd/pull/1814

	result, err := b.mainClient.RawRequest("getblockchaininfo", nil)
	if err != nil {
		return nil, err
	}

	var blockChainInfo types.BlockChainInfo
	if err := json.Unmarshal(result, &blockChainInfo); err != nil {
		return nil, err
	}

	return &blockChainInfo, nil
}
