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

// Package oracle provides an extensible framework for identifying oracle-related
// transactions in the transaction pool (e.g., Chainlink, Redstone price feed updates).
package oracle

import "github.com/ethereum/go-ethereum/core/types"

// OracleType represents the type of oracle provider.
type OracleType string

const (
	OracleChainlink OracleType = "chainlink"
	OracleRedstone  OracleType = "redstone"
)

// OracleInfo contains metadata about an identified oracle transaction.
type OracleInfo struct {
	Type OracleType
	// Extensible: future fields for decoded price data, feed address, etc.
}

// Identifier defines the interface for oracle transaction identification.
// Implementations should check whether a transaction is related to a specific
// oracle provider and return metadata if it matches.
type Identifier interface {
	// Type returns the oracle type this identifier handles.
	Type() OracleType

	// Identify checks whether the given transaction is an oracle transaction.
	// Returns non-nil OracleInfo if the transaction matches, nil otherwise.
	Identify(tx *types.Transaction) *OracleInfo
}
