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
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

var (
	// Redstone updateDataFeedsValues method selector.
	redstoneSelector = [4]byte{0xb7, 0xa1, 0x62, 0x51}

	// Redstone PriceFeedAdapter contract address.
	redstoneTarget = common.HexToAddress("0x97c19d3Ae8e4d74e25EF3AFf3a277fB614ed76D4")
)

// RedstoneIdentifier identifies Redstone oracle transactions.
type RedstoneIdentifier struct{}

// NewRedstoneIdentifier creates a new Redstone oracle identifier.
func NewRedstoneIdentifier() *RedstoneIdentifier {
	return &RedstoneIdentifier{}
}

// Type returns the oracle type for Redstone.
func (r *RedstoneIdentifier) Type() OracleType {
	return OracleRedstone
}

// Identify checks whether the transaction is a Redstone oracle transaction.
func (r *RedstoneIdentifier) Identify(tx *types.Transaction) *OracleInfo {
	// Check single target address first (cheapest filter for single-target oracle).
	to := tx.To()
	if to == nil || *to != redstoneTarget {
		return nil
	}
	// Check method selector.
	data := tx.Data()
	if len(data) < 4 || [4]byte(data[:4]) != redstoneSelector {
		return nil
	}
	return &OracleInfo{Type: OracleRedstone}
}
