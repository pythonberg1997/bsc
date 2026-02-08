package legacypool

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func TestFilterHighGasTxs_AboveThreshold(t *testing.T) {
	addr := common.HexToAddress("0x1234")
	cfg := &HighGasTxConfig{
		Threshold: big.NewInt(1000), // gasCost > 1000
	}
	// gasPrice=10, gasLimit=200 → gasCost=2000 > 1000 ✓
	tx := types.NewTx(&types.LegacyTx{GasPrice: big.NewInt(10), Gas: 200, To: &addr})
	matched := filterHighGasTxs([]*types.Transaction{tx}, cfg)
	if len(matched) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matched))
	}
}

func TestFilterHighGasTxs_BelowThreshold(t *testing.T) {
	addr := common.HexToAddress("0x1234")
	cfg := &HighGasTxConfig{
		Threshold: big.NewInt(1000),
	}
	// gasPrice=1, gasLimit=100 → gasCost=100 <= 1000 ✗
	tx := types.NewTx(&types.LegacyTx{GasPrice: big.NewInt(1), Gas: 100, To: &addr})
	matched := filterHighGasTxs([]*types.Transaction{tx}, cfg)
	if len(matched) != 0 {
		t.Fatalf("expected 0 matches, got %d", len(matched))
	}
}

func TestFilterHighGasTxs_EqualThreshold(t *testing.T) {
	addr := common.HexToAddress("0x1234")
	cfg := &HighGasTxConfig{
		Threshold: big.NewInt(1000),
	}
	// gasPrice=10, gasLimit=100 → gasCost=1000 == threshold → NOT matched (strictly greater)
	tx := types.NewTx(&types.LegacyTx{GasPrice: big.NewInt(10), Gas: 100, To: &addr})
	matched := filterHighGasTxs([]*types.Transaction{tx}, cfg)
	if len(matched) != 0 {
		t.Fatalf("expected 0 matches for equal threshold, got %d", len(matched))
	}
}

func TestFilterHighGasTxs_Whitelisted(t *testing.T) {
	addr := common.HexToAddress("0xAAAA")
	cfg := &HighGasTxConfig{
		Threshold: big.NewInt(100),
		Whitelist: map[common.Address]struct{}{
			addr: {},
		},
	}
	// Above threshold but To address is whitelisted → excluded
	tx := types.NewTx(&types.LegacyTx{GasPrice: big.NewInt(10), Gas: 200, To: &addr})
	matched := filterHighGasTxs([]*types.Transaction{tx}, cfg)
	if len(matched) != 0 {
		t.Fatalf("expected 0 matches for whitelisted address, got %d", len(matched))
	}
}

func TestFilterHighGasTxs_NotWhitelisted(t *testing.T) {
	whitelisted := common.HexToAddress("0xAAAA")
	target := common.HexToAddress("0xBBBB")
	cfg := &HighGasTxConfig{
		Threshold: big.NewInt(100),
		Whitelist: map[common.Address]struct{}{
			whitelisted: {},
		},
	}
	// Above threshold and To address is NOT whitelisted → matched
	tx := types.NewTx(&types.LegacyTx{GasPrice: big.NewInt(10), Gas: 200, To: &target})
	matched := filterHighGasTxs([]*types.Transaction{tx}, cfg)
	if len(matched) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matched))
	}
}

func TestFilterHighGasTxs_NilTo(t *testing.T) {
	cfg := &HighGasTxConfig{
		Threshold: big.NewInt(100),
		Whitelist: map[common.Address]struct{}{
			common.HexToAddress("0xAAAA"): {},
		},
	}
	// Contract creation (nil To) above threshold → matched (whitelist doesn't apply)
	tx := types.NewTx(&types.LegacyTx{GasPrice: big.NewInt(10), Gas: 200})
	matched := filterHighGasTxs([]*types.Transaction{tx}, cfg)
	if len(matched) != 1 {
		t.Fatalf("expected 1 match for nil To, got %d", len(matched))
	}
}

func TestFilterHighGasTxs_GasLimitCap(t *testing.T) {
	addr := common.HexToAddress("0x1234")
	cfg := &HighGasTxConfig{
		Threshold:   big.NewInt(100),
		GasLimitCap: 500, // Only accept txs with gasLimit < 500
	}

	// gasLimit=200 < 500 and gasCost=2000 > 100 → matched
	tx1 := types.NewTx(&types.LegacyTx{GasPrice: big.NewInt(10), Gas: 200, To: &addr})
	// gasLimit=500 >= 500 → excluded by GasLimitCap
	tx2 := types.NewTx(&types.LegacyTx{GasPrice: big.NewInt(10), Gas: 500, To: &addr})
	// gasLimit=1000 >= 500 → excluded by GasLimitCap
	tx3 := types.NewTx(&types.LegacyTx{GasPrice: big.NewInt(10), Gas: 1000, To: &addr})

	matched := filterHighGasTxs([]*types.Transaction{tx1, tx2, tx3}, cfg)
	if len(matched) != 1 {
		t.Fatalf("expected 1 match with gasLimitCap, got %d", len(matched))
	}
}

func TestFilterHighGasTxs_GasLimitCapZero(t *testing.T) {
	addr := common.HexToAddress("0x1234")
	cfg := &HighGasTxConfig{
		Threshold:   big.NewInt(100),
		GasLimitCap: 0, // 0 = no cap
	}
	// GasLimitCap=0 means no cap, so large gasLimit is fine
	tx := types.NewTx(&types.LegacyTx{GasPrice: big.NewInt(10), Gas: 999999, To: &addr})
	matched := filterHighGasTxs([]*types.Transaction{tx}, cfg)
	if len(matched) != 1 {
		t.Fatalf("expected 1 match with zero gasLimitCap, got %d", len(matched))
	}
}

func TestFilterHighGasTxs_EIP1559(t *testing.T) {
	addr := common.HexToAddress("0x1234")
	cfg := &HighGasTxConfig{
		Threshold: big.NewInt(1000),
	}
	// EIP-1559 tx: GasFeeCap=20, GasTipCap=5, Gas=100 → gasCost = 20*100 = 2000 > 1000 ✓
	tx := types.NewTx(&types.DynamicFeeTx{
		GasFeeCap: big.NewInt(20),
		GasTipCap: big.NewInt(5),
		Gas:       100,
		To:        &addr,
	})
	matched := filterHighGasTxs([]*types.Transaction{tx}, cfg)
	if len(matched) != 1 {
		t.Fatalf("expected 1 match for EIP-1559 tx, got %d", len(matched))
	}
}

func TestFilterHighGasTxs_MixedBatch(t *testing.T) {
	addr := common.HexToAddress("0x1234")
	whitelisted := common.HexToAddress("0xAAAA")
	cfg := &HighGasTxConfig{
		Threshold:   big.NewInt(500),
		GasLimitCap: 1000,
		Whitelist: map[common.Address]struct{}{
			whitelisted: {},
		},
	}

	txs := []*types.Transaction{
		// gasCost=2000 > 500, gasLimit=200 < 1000, not whitelisted → ✓
		types.NewTx(&types.LegacyTx{GasPrice: big.NewInt(10), Gas: 200, To: &addr}),
		// gasCost=100 <= 500 → ✗ (below threshold)
		types.NewTx(&types.LegacyTx{GasPrice: big.NewInt(1), Gas: 100, To: &addr}),
		// gasCost=20000 > 500, gasLimit=2000 >= 1000 → ✗ (exceeds gasLimitCap)
		types.NewTx(&types.LegacyTx{GasPrice: big.NewInt(10), Gas: 2000, To: &addr}),
		// gasCost=2000 > 500, gasLimit=200 < 1000, but whitelisted → ✗
		types.NewTx(&types.LegacyTx{GasPrice: big.NewInt(10), Gas: 200, To: &whitelisted}),
		// gasCost=2000 > 500, gasLimit=200 < 1000, nil To → ✓
		types.NewTx(&types.LegacyTx{GasPrice: big.NewInt(10), Gas: 200}),
	}

	matched := filterHighGasTxs(txs, cfg)
	if len(matched) != 2 {
		t.Fatalf("expected 2 matches in mixed batch, got %d", len(matched))
	}
}

func TestFilterHighGasTxs_EmptyWhitelist(t *testing.T) {
	addr := common.HexToAddress("0x1234")
	cfg := &HighGasTxConfig{
		Threshold: big.NewInt(100),
		Whitelist: nil, // nil whitelist
	}
	tx := types.NewTx(&types.LegacyTx{GasPrice: big.NewInt(10), Gas: 200, To: &addr})
	matched := filterHighGasTxs([]*types.Transaction{tx}, cfg)
	if len(matched) != 1 {
		t.Fatalf("expected 1 match with nil whitelist, got %d", len(matched))
	}
}
