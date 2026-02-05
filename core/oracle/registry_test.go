package oracle

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// mockIdentifier is a test helper that matches transactions with a specific nonce.
type mockIdentifier struct {
	oracleType OracleType
	matchNonce uint64
}

func (m *mockIdentifier) Type() OracleType { return m.oracleType }

func (m *mockIdentifier) Identify(tx *types.Transaction) *OracleInfo {
	if tx.Nonce() == m.matchNonce {
		return &OracleInfo{Type: m.oracleType}
	}
	return nil
}

func TestRegistryEmpty(t *testing.T) {
	r := NewRegistry()
	tx := types.NewTx(&types.LegacyTx{Nonce: 1})
	if info := r.Identify(tx); info != nil {
		t.Fatalf("empty registry should return nil, got %v", info)
	}
	if r.Len() != 0 {
		t.Fatalf("empty registry length should be 0, got %d", r.Len())
	}
}

func TestRegistryFirstMatch(t *testing.T) {
	r := NewRegistry()
	// Both identifiers match nonce 42, first one should win.
	r.Register(&mockIdentifier{oracleType: "oracle_a", matchNonce: 42})
	r.Register(&mockIdentifier{oracleType: "oracle_b", matchNonce: 42})

	tx := types.NewTx(&types.LegacyTx{Nonce: 42})
	info := r.Identify(tx)
	if info == nil {
		t.Fatal("expected match, got nil")
	}
	if info.Type != "oracle_a" {
		t.Fatalf("expected first-match oracle_a, got %s", info.Type)
	}
}

func TestRegistryNoMatch(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockIdentifier{oracleType: OracleChainlink, matchNonce: 100})
	r.Register(&mockIdentifier{oracleType: OracleRedstone, matchNonce: 200})

	tx := types.NewTx(&types.LegacyTx{Nonce: 999})
	if info := r.Identify(tx); info != nil {
		t.Fatalf("expected no match, got %v", info)
	}
}

func TestRegistrySelectiveMatch(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockIdentifier{oracleType: OracleChainlink, matchNonce: 10})
	r.Register(&mockIdentifier{oracleType: OracleRedstone, matchNonce: 20})

	// Only second identifier should match.
	tx := types.NewTx(&types.LegacyTx{Nonce: 20})
	info := r.Identify(tx)
	if info == nil {
		t.Fatal("expected match, got nil")
	}
	if info.Type != OracleRedstone {
		t.Fatalf("expected redstone, got %s", info.Type)
	}
}

func TestRegistryLen(t *testing.T) {
	r := NewRegistry()
	if r.Len() != 0 {
		t.Fatalf("expected 0, got %d", r.Len())
	}
	r.Register(NewChainlinkIdentifier())
	r.Register(NewRedstoneIdentifier())
	if r.Len() != 2 {
		t.Fatalf("expected 2, got %d", r.Len())
	}
}

func TestChainlinkIdentify(t *testing.T) {
	// Inject a test target address.
	addr := common.HexToAddress("0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	chainlinkTargets[addr] = struct{}{}
	defer delete(chainlinkTargets, addr)

	cl := NewChainlinkIdentifier()
	if cl.Type() != OracleChainlink {
		t.Fatalf("expected %s, got %s", OracleChainlink, cl.Type())
	}

	// Match with first selector.
	data := append(chainlinkSelectors[0][:], make([]byte, 32)...)
	tx := types.NewTx(&types.LegacyTx{Data: data, To: &addr})
	info := cl.Identify(tx)
	if info == nil || info.Type != OracleChainlink {
		t.Fatalf("expected chainlink match with selector 0, got %v", info)
	}

	// Match with second selector.
	data = append(chainlinkSelectors[1][:], make([]byte, 32)...)
	tx = types.NewTx(&types.LegacyTx{Data: data, To: &addr})
	info = cl.Identify(tx)
	if info == nil || info.Type != OracleChainlink {
		t.Fatalf("expected chainlink match with selector 1, got %v", info)
	}

	// Wrong selector.
	badData := append([]byte{0x00, 0x00, 0x00, 0x00}, make([]byte, 32)...)
	tx = types.NewTx(&types.LegacyTx{Data: badData, To: &addr})
	if cl.Identify(tx) != nil {
		t.Fatal("expected nil for wrong selector")
	}

	// Wrong address.
	unknownAddr := common.HexToAddress("0xdead")
	tx = types.NewTx(&types.LegacyTx{Data: data, To: &unknownAddr})
	if cl.Identify(tx) != nil {
		t.Fatal("expected nil for unknown address")
	}

	// No data.
	tx = types.NewTx(&types.LegacyTx{To: &addr})
	if cl.Identify(tx) != nil {
		t.Fatal("expected nil for empty data")
	}

	// Nil To (contract creation).
	tx = types.NewTx(&types.LegacyTx{Data: data})
	if cl.Identify(tx) != nil {
		t.Fatal("expected nil for nil To")
	}
}

func TestRedstoneIdentify(t *testing.T) {
	// Inject a test target address.
	target := common.HexToAddress("0xBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB")
	origTarget := redstoneTarget
	redstoneTarget = target
	defer func() { redstoneTarget = origTarget }()

	rs := NewRedstoneIdentifier()
	if rs.Type() != OracleRedstone {
		t.Fatalf("expected %s, got %s", OracleRedstone, rs.Type())
	}

	// Matching tx.
	data := append(redstoneSelector[:], make([]byte, 32)...)
	tx := types.NewTx(&types.LegacyTx{Data: data, To: &target})
	info := rs.Identify(tx)
	if info == nil || info.Type != OracleRedstone {
		t.Fatalf("expected redstone match, got %v", info)
	}

	// Wrong address.
	wrongAddr := common.HexToAddress("0xdead")
	tx = types.NewTx(&types.LegacyTx{Data: data, To: &wrongAddr})
	if rs.Identify(tx) != nil {
		t.Fatal("expected nil for wrong address")
	}

	// Wrong selector.
	badData := append([]byte{0x00, 0x00, 0x00, 0x00}, make([]byte, 32)...)
	tx = types.NewTx(&types.LegacyTx{Data: badData, To: &target})
	if rs.Identify(tx) != nil {
		t.Fatal("expected nil for wrong selector")
	}

	// No data.
	tx = types.NewTx(&types.LegacyTx{To: &target})
	if rs.Identify(tx) != nil {
		t.Fatal("expected nil for empty data")
	}

	// Nil To.
	tx = types.NewTx(&types.LegacyTx{Data: data})
	if rs.Identify(tx) != nil {
		t.Fatal("expected nil for nil To")
	}
}
