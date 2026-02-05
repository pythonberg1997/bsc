// Copyright 2026 The bsc Authors
// This file is part of the bsc library.
//
// The bsc library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The bsc library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the bsc library. If not, see <http://www.gnu.org/licenses/>.
package oracle

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	chainlinkFactoryBSC = "0x297Bc37BCE59112D245cCfC22f54079466733401"
)

// TestScanChainlinkForwarders scans AuthorizedForwarderCreated events from the
// Chainlink OperatorFactory and prints unique forwarder addresses.
//
// Env vars:
// - BSC_RPC_URL: required RPC endpoint (e.g., https://bsc-dataseed.binance.org)
// - FROM_BLOCK: optional start block (default 0)
// - TO_BLOCK: optional end block (default latest)
// - BATCH_SIZE: optional batch size (default 50000)
func TestScanChainlinkForwarders(t *testing.T) {
	rpcURL := "http://127.0.0.1:8544"

	fromBlock := uint64(40299701) // deployed at block 40299701
	toBlock := uint64(44844259)   // known latest tx height
	batchSize := uint64(50000)

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		t.Fatalf("dial rpc: %v", err)
	}
	defer client.Close()

	if toBlock == 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		latest, err := client.BlockNumber(ctx)
		if err != nil {
			t.Fatalf("get latest block: %v", err)
		}
		toBlock = latest
	}

	factoryAddr := common.HexToAddress(chainlinkFactoryBSC)
	eventSig := crypto.Keccak256Hash([]byte("AuthorizedForwarderCreated(address,address,address)"))

	forwarders := map[common.Address]struct{}{}
	var logsCount int

	for start := fromBlock; start <= toBlock; start += batchSize {
		end := start + batchSize - 1
		if end > toBlock {
			end = toBlock
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		query := ethereum.FilterQuery{
			FromBlock: bigInt(start),
			ToBlock:   bigInt(end),
			Addresses: []common.Address{factoryAddr},
			Topics:    [][]common.Hash{{eventSig}},
		}
		logs, err := client.FilterLogs(ctx, query)
		cancel()
		if err != nil {
			t.Fatalf("filter logs [%d-%d]: %v", start, end, err)
		}

		for _, log := range logs {
			logsCount++
			fwd, owner, sender, ok := decodeForwarderEvent(log)
			if !ok {
				continue
			}
			forwarders[fwd] = struct{}{}
			t.Logf("forwarder=%s owner=%s sender=%s block=%d tx=%s",
				fwd.Hex(), owner.Hex(), sender.Hex(), log.BlockNumber, log.TxHash.Hex())
		}
	}

	list := make([]string, 0, len(forwarders))
	for addr := range forwarders {
		list = append(list, addr.Hex())
	}
	sort.Strings(list)

	t.Logf("total logs=%d unique forwarders=%d", logsCount, len(list))
	for _, addr := range list {
		fmt.Println(addr)
	}
}

func decodeForwarderEvent(log types.Log) (common.Address, common.Address, common.Address, bool) {
	// Topics: [sig, forwarder, owner, sender]
	if len(log.Topics) < 4 {
		return common.Address{}, common.Address{}, common.Address{}, false
	}
	return topicToAddress(log.Topics[1]),
		topicToAddress(log.Topics[2]),
		topicToAddress(log.Topics[3]),
		true
}

func topicToAddress(topic common.Hash) common.Address {
	b := topic.Bytes()
	return common.BytesToAddress(b[12:])
}

func parseUintEnv(t *testing.T, key string, def uint64) uint64 {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		t.Fatalf("invalid %s=%q: %v", key, v, err)
	}
	return n
}

func bigInt(v uint64) *big.Int {
	return new(big.Int).SetUint64(v)
}
