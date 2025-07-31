package svc

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/ledgerhq/satstack/bus"
	"github.com/ledgerhq/satstack/version"
	log "github.com/sirupsen/logrus"
)

func (s *Service) GetHealth() error {
	// Custom blockchain info struct to avoid btcd struct incompatibility
	type customBlockChainInfo struct {
		Warnings []string `json:"warnings"`
	}

	client, err := s.Bus.ClientFactory()
	if err != nil {
		return err
	}
	defer client.Shutdown()

	result, err := client.RawRequest("getblockchaininfo", nil)
	if err != nil {
		return err
	}

	var info customBlockChainInfo
	if err := json.Unmarshal(result, &info); err != nil {
		return fmt.Errorf("unable to parse blockchain info: %w", err)
	}

	// TODO: Check contents of GetBlockChainInfo response

	return nil
}

func (s *Service) GetFees(targets []int64, mode string) map[string]interface{} {
	result := make(map[string]interface{})
	for _, target := range targets {
		fee := s.Bus.EstimateSmartFee(target, mode)
		result[strconv.FormatInt(target, 10)] = fee
	}

	result["last_updated"] = int32(time.Now().Unix())
	return result
}

func (s *Service) GetStatus() *bus.ExplorerStatus {
	// Prepare base bus.ExplorerStatus instance.
	status := bus.ExplorerStatus{
		Version:  version.Version,
		TxIndex:  s.Bus.TxIndex,
		Pruned:   s.Bus.Pruned,
		Chain:    s.Bus.Chain,
		Currency: s.Bus.Currency,
	}

	// Case 1: satstack is running the numbers.
	// or rescanning the wallet
	if s.Bus.IsPendingScan {
		status.Status = bus.PendingScan
		return &status
	}

	// Case 2: Unable to initialize rpcclient.Client.
	client, err := s.Bus.ClientFactory()
	if err != nil {
		log.WithField(
			"err", fmt.Errorf("%s: %w", bus.ErrBitcoindUnreachable, err),
		).Error("Failed to query status")
		status.Status = bus.NodeDisconnected
		return &status
	}

	defer client.Shutdown()

	// Case 3: bitcoind is unreachable - chain RPC failed.
	// Custom blockchain info struct to avoid btcd struct incompatibility
	type customBlockChainInfo struct {
		Blocks               int32   `json:"blocks"`
		Headers              int32   `json:"headers"`
		VerificationProgress float64 `json:"verificationprogress"`
		Warnings             []string `json:"warnings"`
	}

	result, err := client.RawRequest("getblockchaininfo", nil)
	if err != nil {
		log.WithField(
			"err", fmt.Errorf("%s: %w", bus.ErrBitcoindUnreachable, err),
		).Error("Failed to query status")

		status.Status = bus.NodeDisconnected
		return &status
	}

	var blockChainInfo customBlockChainInfo
	if err := json.Unmarshal(result, &blockChainInfo); err != nil {
		log.WithField(
			"err", fmt.Errorf("unable to parse blockchain info: %w", err),
		).Error("Failed to query status")

		status.Status = bus.NodeDisconnected
		return &status
	}

	// Case 4: bitcoind is currently catching up on new blocks.
	if blockChainInfo.Blocks != blockChainInfo.Headers {
		status.Status = bus.Syncing
		status.SyncProgress = btcjson.Float64(
			blockChainInfo.VerificationProgress * 100)
		return &status
	}

	// Case 5: bitcoind is currently importing descriptors
	walletInfo, err := client.GetWalletInfo()
	if err != nil {
		log.WithField(
			"err", fmt.Errorf("%s: %w", bus.ErrBitcoindUnreachable, err),
		).Error("Failed to query status")

		status.Status = bus.NodeDisconnected
		return &status
	}

	switch v := walletInfo.Scanning.Value.(type) {
	case btcjson.ScanProgress:
		status.Status = bus.Scanning
		status.ScanProgress = btcjson.Float64(v.Progress * 100)
		return &status
	}

	// Case 6: bitcoind is ready to be used with satstack.
	status.Status = bus.Ready
	return &status
}

func (s *Service) GetNetwork() (network *bus.Network) {
	client, err := s.Bus.ClientFactory()
	if err != nil {
		log.WithField("err", fmt.Errorf("%s: %w", bus.ErrBitcoindUnreachable, err)).
			Error("Failed to query status")

		network = new(bus.Network)
		return network
	}

	// Custom network info struct to handle warnings as array
	type customNetworkInfo struct {
		RelayFee       float64  `json:"relayfee"`
		IncrementalFee float64  `json:"incrementalfee"`
		Version        int32    `json:"version"`
		Subversion     string   `json:"subversion"`
		Warnings       []string `json:"warnings"`
	}

	// Use raw request to avoid btcd struct incompatibility
	result, err := client.RawRequest("getnetworkinfo", nil)
	if err != nil {
		log.WithField("err", fmt.Errorf("%s: %w", bus.ErrBitcoindUnreachable, err)).
			Error("Failed to query status")

		network = new(bus.Network)
		return network
	}

	var networkInfo customNetworkInfo
	if err := json.Unmarshal(result, &networkInfo); err != nil {
		log.WithField("err", fmt.Errorf("unable to parse network info: %w", err)).
			Error("Failed to query status")

		network = new(bus.Network)
		return network
	}

	network = &bus.Network{
		RelayFee:       networkInfo.RelayFee,
		IncrementalFee: networkInfo.IncrementalFee,
		Version:        networkInfo.Version,
		Subversion:     networkInfo.Subversion,
	}
	return network
}
