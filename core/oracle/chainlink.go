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
	// Chainlink method selectors.
	chainlinkSelectors = [4][4]byte{
		{0x6f, 0xad, 0xcf, 0x72},
		{0xb6, 0x4f, 0xa9, 0xe6},
		{0xb1, 0xdc, 0x65, 0xa4}, // dual transmit
		{0xba, 0x0c, 0xb2, 0x9e}, // dual transmitSecondary
	}

	// Chainlink forwarder contract addresses.
	chainlinkTargets = map[common.Address]struct{}{
		common.HexToAddress("0x080f02795ba9003404Aceaf102f14443647DBe08"): {},
		common.HexToAddress("0x1877d26Dabef3b5869aAA788EbFB39F1d6C477A0"): {},
		common.HexToAddress("0x1b66fBd3a0c4aa37705E7091EF40554E124A25F6"): {},
		common.HexToAddress("0x2c14565cDDC53F3a60D27E8C345809a31D934642"): {},
		common.HexToAddress("0x328e4BDcf3096C36D9CB2d989Da6995b461140ac"): {},
		common.HexToAddress("0x342682197497FC9b860e914636847F1a9AD8b037"): {},
		common.HexToAddress("0x347dF906a432D7964Ce2D854B51D9166B6A46CcC"): {},
		common.HexToAddress("0x3A642243e38beA40322aa9A7e59c26Aa27Cc6F8D"): {},
		common.HexToAddress("0x42b950F1Cf4855a54f6d0bd6981687Eb4d5d6a18"): {},
		common.HexToAddress("0x5b4813eE2D4366D73C89D7870aFa82BF2FBB3E71"): {},
		common.HexToAddress("0x5d6173E05c3Af359DBb13ee12Af726E34AAeDAD2"): {},
		common.HexToAddress("0x64F4d51eE8AE9671c5d1cE58dc15a653A71efC90"): {},
		common.HexToAddress("0x7511839dFfaf432A5DbA567bc1Aa77115EfEf882"): {},
		common.HexToAddress("0x870A3D3BBB5554f1D04967a6b76534989020B203"): {},
		common.HexToAddress("0x8bDB5E089E9285037E6874D8F3160b5548611380"): {},
		common.HexToAddress("0x9A7032C6EDCBbf164E201ca9dF5894BC35416e6A"): {},
		common.HexToAddress("0x9dFCA5435036ad0069ECa7C8e15b0c5c44548952"): {},
		common.HexToAddress("0xD979873cFAf9a655F44006Fe01B148f1ABbc47Cc"): {},
		common.HexToAddress("0xF72bb66E324B5188c6178c6401Ee982FA30bE096"): {},
		common.HexToAddress("0xF87C2E9f08c04f805d15C5E90273227639151860"): {},
		common.HexToAddress("0xa13814fE399ea85e0541B4FdD2b38efa149c8Ee5"): {},
		common.HexToAddress("0xaaC815bcd3d1eCEACfC51608c4EaF194eEfd14f1"): {},
		common.HexToAddress("0xc3bC83292fcc9c646E61f37126839449736d64FD"): {},
		common.HexToAddress("0xc7F79eA9c7FF17968eEC60b4CC9880905cFdb2f4"): {},
		common.HexToAddress("0xdB1A1C98F4A23c8C78B7B2c44Cd9678Ff286A0A1"): {},
	}
)

// ChainlinkIdentifier identifies Chainlink oracle transactions.
type ChainlinkIdentifier struct{}

// NewChainlinkIdentifier creates a new Chainlink oracle identifier.
func NewChainlinkIdentifier() *ChainlinkIdentifier {
	return &ChainlinkIdentifier{}
}

// Type returns the oracle type for Chainlink.
func (c *ChainlinkIdentifier) Type() OracleType {
	return OracleChainlink
}

// Identify checks whether the transaction is a Chainlink oracle transaction.
func (c *ChainlinkIdentifier) Identify(tx *types.Transaction) *OracleInfo {
	// Check method selector (4-byte comparison, very cheap).
	data := tx.Data()
	if len(data) < 4 {
		return nil
	}
	sel := [4]byte(data[:4])
	if sel != chainlinkSelectors[0] && sel != chainlinkSelectors[1] &&
		sel != chainlinkSelectors[2] && sel != chainlinkSelectors[3] {
		return nil
	}
	// Check target address against the known set.
	to := tx.To()
	if to == nil {
		return nil
	}
	if _, ok := chainlinkTargets[*to]; !ok {
		return nil
	}
	return &OracleInfo{Type: OracleChainlink}
}
