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

import "github.com/ethereum/go-ethereum/core/types"

// Registry manages a collection of oracle identifiers and provides
// a unified interface to classify transactions.
type Registry struct {
	identifiers []Identifier
}

// NewRegistry creates an empty oracle registry.
func NewRegistry() *Registry {
	return &Registry{}
}

// Register adds an oracle identifier to the registry.
func (r *Registry) Register(id Identifier) {
	r.identifiers = append(r.identifiers, id)
}

// Identify checks the transaction against all registered identifiers.
// Returns the first match (first-match semantics). Returns nil if no
// identifier recognizes the transaction.
func (r *Registry) Identify(tx *types.Transaction) *OracleInfo {
	for _, id := range r.identifiers {
		if info := id.Identify(tx); info != nil {
			return info
		}
	}
	return nil
}

// Len returns the number of registered identifiers.
func (r *Registry) Len() int {
	return len(r.identifiers)
}
