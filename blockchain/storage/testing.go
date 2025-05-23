package storage

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/apex-fusion/nexus/helper/hex"
	"github.com/apex-fusion/nexus/types"
	"github.com/stretchr/testify/assert"
)

type PlaceholderStorage func(t *testing.T) (Storage, func())

var (
	addr1 = types.StringToAddress("1")
	addr2 = types.StringToAddress("2")

	hash1 = types.StringToHash("1")
	hash2 = types.StringToHash("2")
)

// TestStorage tests a set of tests on a storage
func TestStorage(t *testing.T, m PlaceholderStorage) {
	t.Helper()

	t.Run("", func(t *testing.T) {
		testCanonicalChain(t, m)
	})
	t.Run("", func(t *testing.T) {
		testDifficulty(t, m)
	})
	t.Run("", func(t *testing.T) {
		testHead(t, m)
	})
	t.Run("", func(t *testing.T) {
		testForks(t, m)
	})
	t.Run("", func(t *testing.T) {
		testHeader(t, m)
	})
	t.Run("", func(t *testing.T) {
		testWriteCanonicalHeader(t, m)
	})
}

func testCanonicalChain(t *testing.T, m PlaceholderStorage) {
	t.Helper()

	s, closeFn := m(t)
	defer closeFn()

	var cases = []struct {
		Number     uint64
		ParentHash types.Hash
		Hash       types.Hash
	}{
		{
			Number:     1,
			ParentHash: types.StringToHash("111"),
		},
		{
			Number:     1,
			ParentHash: types.StringToHash("222"),
		},
		{
			Number:     2,
			ParentHash: types.StringToHash("111"),
		},
	}

	for _, cc := range cases {
		h := &types.Header{
			Number:     cc.Number,
			ParentHash: cc.ParentHash,
			ExtraData:  []byte{0x1},
		}

		hash := h.Hash

		if err := s.WriteHeader(h); err != nil {
			t.Fatal(err)
		}

		if err := s.WriteCanonicalHash(cc.Number, hash); err != nil {
			t.Fatal(err)
		}

		data, ok := s.ReadCanonicalHash(cc.Number)
		if !ok {
			t.Fatal("not found")
		}

		if !reflect.DeepEqual(data, hash) {
			t.Fatal("not match")
		}
	}
}

func testDifficulty(t *testing.T, m PlaceholderStorage) {
	t.Helper()

	s, closeFn := m(t)
	defer closeFn()

	var cases = []struct {
		Diff *big.Int
	}{
		{
			Diff: big.NewInt(10),
		},
		{
			Diff: big.NewInt(11),
		},
		{
			Diff: big.NewInt(12),
		},
	}

	for indx, cc := range cases {
		h := &types.Header{
			Number:    uint64(indx),
			ExtraData: []byte{},
		}

		hash := h.Hash

		if err := s.WriteHeader(h); err != nil {
			t.Fatal(err)
		}

		if err := s.WriteTotalDifficulty(hash, cc.Diff); err != nil {
			t.Fatal(err)
		}

		diff, ok := s.ReadTotalDifficulty(hash)
		if !ok {
			t.Fatal("not found")
		}

		if !reflect.DeepEqual(cc.Diff, diff) {
			t.Fatal("bad")
		}
	}
}

func testHead(t *testing.T, m PlaceholderStorage) {
	t.Helper()

	s, closeFn := m(t)
	defer closeFn()

	for i := uint64(0); i < 5; i++ {
		h := &types.Header{
			Number:    i,
			ExtraData: []byte{},
		}
		hash := h.Hash

		if err := s.WriteHeader(h); err != nil {
			t.Fatal(err)
		}

		if err := s.WriteHeadNumber(i); err != nil {
			t.Fatal(err)
		}

		if err := s.WriteHeadHash(hash); err != nil {
			t.Fatal(err)
		}

		n2, ok := s.ReadHeadNumber()
		if !ok {
			t.Fatal("num not found")
		}

		if n2 != i {
			t.Fatal("bad")
		}

		hash1, ok := s.ReadHeadHash()
		if !ok {
			t.Fatal("hash not found")
		}

		if !reflect.DeepEqual(hash1, hash) {
			t.Fatal("bad")
		}
	}
}

func testForks(t *testing.T, m PlaceholderStorage) {
	t.Helper()

	s, closeFn := m(t)
	defer closeFn()

	var cases = []struct {
		Forks []types.Hash
	}{
		{[]types.Hash{types.StringToHash("111"), types.StringToHash("222")}},
		{[]types.Hash{types.StringToHash("111")}},
	}

	for _, cc := range cases {
		if err := s.WriteForks(cc.Forks); err != nil {
			t.Fatal(err)
		}

		forks, err := s.ReadForks()
		assert.NoError(t, err)

		if !reflect.DeepEqual(cc.Forks, forks) {
			t.Fatal("bad")
		}
	}
}

func testHeader(t *testing.T, m PlaceholderStorage) {
	t.Helper()

	s, closeFn := m(t)
	defer closeFn()

	header := &types.Header{
		Number:     5,
		Difficulty: 17179869184,
		ParentHash: types.StringToHash("11"),
		Timestamp:  10,
		// if not set it will fail
		ExtraData: hex.MustDecodeHex("0x11bbe8db4e347b4e8c937c1c8370e4b5ed33adb3db69cbdb7a38e1e50b1b82fa"),
	}
	header.ComputeHash()

	if err := s.WriteHeader(header); err != nil {
		t.Fatal(err)
	}

	header1, err := s.ReadHeader(header.Hash)
	assert.NoError(t, err)

	if !reflect.DeepEqual(header, header1) {
		t.Fatal("bad")
	}
}

func testWriteCanonicalHeader(t *testing.T, m PlaceholderStorage) {
	t.Helper()

	s, closeFn := m(t)
	defer closeFn()

	h := &types.Header{
		Number:    100,
		ExtraData: []byte{0x1},
	}
	h.ComputeHash()

	diff := new(big.Int).SetUint64(100)

	if err := s.WriteCanonicalHeader(h, diff); err != nil {
		t.Fatal(err)
	}

	hh, err := s.ReadHeader(h.Hash)
	assert.NoError(t, err)

	if !reflect.DeepEqual(h, hh) {
		t.Fatal("bad header")
	}

	headHash, ok := s.ReadHeadHash()
	if !ok {
		t.Fatal("not found head hash")
	}

	if headHash != h.Hash {
		t.Fatal("head hash not correct")
	}

	headNum, ok := s.ReadHeadNumber()
	if !ok {
		t.Fatal("not found head num")
	}

	if headNum != h.Number {
		t.Fatal("head num not correct")
	}

	canHash, ok := s.ReadCanonicalHash(h.Number)
	if !ok {
		t.Fatal("not found can hash")
	}

	if canHash != h.Hash {
		t.Fatal("canonical hash not correct")
	}
}

// Storage delegators

type readCanonicalHashDelegate func(uint64) (types.Hash, bool)
type writeCanonicalHashDelegate func(uint64, types.Hash) error
type readHeadHashDelegate func() (types.Hash, bool)
type readHeadNumberDelegate func() (uint64, bool)
type writeHeadHashDelegate func(types.Hash) error
type writeHeadNumberDelegate func(uint64) error
type writeForksDelegate func([]types.Hash) error
type readForksDelegate func() ([]types.Hash, error)
type writeTotalDifficultyDelegate func(types.Hash, *big.Int) error
type readTotalDifficultyDelegate func(types.Hash) (*big.Int, bool)
type writeHeaderDelegate func(*types.Header) error
type readHeaderDelegate func(types.Hash) (*types.Header, error)
type writeCanonicalHeaderDelegate func(*types.Header, *big.Int) error
type writeBodyDelegate func(types.Hash, *types.Body) error
type readBodyDelegate func(types.Hash) (*types.Body, error)
type writeSnapshotDelegate func(types.Hash, []byte) error
type readSnapshotDelegate func(types.Hash) ([]byte, bool)
type writeReceiptsDelegate func(types.Hash, []*types.Receipt) error
type readReceiptsDelegate func(types.Hash) ([]*types.Receipt, error)
type writeTxLookupDelegate func(types.Hash, types.Hash) error
type readTxLookupDelegate func(types.Hash) (types.Hash, bool)
type closeDelegate func() error

type MockStorage struct {
	readCanonicalHashFn    readCanonicalHashDelegate
	writeCanonicalHashFn   writeCanonicalHashDelegate
	readHeadHashFn         readHeadHashDelegate
	readHeadNumberFn       readHeadNumberDelegate
	writeHeadHashFn        writeHeadHashDelegate
	writeHeadNumberFn      writeHeadNumberDelegate
	writeForksFn           writeForksDelegate
	readForksFn            readForksDelegate
	writeTotalDifficultyFn writeTotalDifficultyDelegate
	readTotalDifficultyFn  readTotalDifficultyDelegate
	writeHeaderFn          writeHeaderDelegate
	readHeaderFn           readHeaderDelegate
	writeCanonicalHeaderFn writeCanonicalHeaderDelegate
	writeBodyFn            writeBodyDelegate
	readBodyFn             readBodyDelegate
	writeReceiptsFn        writeReceiptsDelegate
	readReceiptsFn         readReceiptsDelegate
	writeTxLookupFn        writeTxLookupDelegate
	readTxLookupFn         readTxLookupDelegate
	closeFn                closeDelegate
}

func NewMockStorage() *MockStorage {
	return &MockStorage{}
}

func (m *MockStorage) ReadCanonicalHash(n uint64) (types.Hash, bool) {
	if m.readCanonicalHashFn != nil {
		return m.readCanonicalHashFn(n)
	}

	return types.Hash{}, true
}

func (m *MockStorage) HookReadCanonicalHash(fn readCanonicalHashDelegate) {
	m.readCanonicalHashFn = fn
}

func (m *MockStorage) WriteCanonicalHash(n uint64, hash types.Hash) error {
	if m.writeCanonicalHashFn != nil {
		return m.writeCanonicalHashFn(n, hash)
	}

	return nil
}

func (m *MockStorage) HookWriteCanonicalHash(fn writeCanonicalHashDelegate) {
	m.writeCanonicalHashFn = fn
}

func (m *MockStorage) ReadHeadHash() (types.Hash, bool) {
	if m.readHeadHashFn != nil {
		return m.readHeadHashFn()
	}

	return types.Hash{}, true
}

func (m *MockStorage) HookReadHeadHash(fn readHeadHashDelegate) {
	m.readHeadHashFn = fn
}

func (m *MockStorage) ReadHeadNumber() (uint64, bool) {
	if m.readHeadNumberFn != nil {
		return m.readHeadNumberFn()
	}

	return 0, true
}

func (m *MockStorage) HookReadHeadNumber(fn readHeadNumberDelegate) {
	m.readHeadNumberFn = fn
}

func (m *MockStorage) WriteHeadHash(h types.Hash) error {
	if m.writeHeadHashFn != nil {
		return m.writeHeadHashFn(h)
	}

	return nil
}

func (m *MockStorage) HookWriteHeadHash(fn writeHeadHashDelegate) {
	m.writeHeadHashFn = fn
}

func (m *MockStorage) WriteHeadNumber(n uint64) error {
	if m.writeHeadNumberFn != nil {
		return m.writeHeadNumberFn(n)
	}

	return nil
}

func (m *MockStorage) HookWriteHeadNumber(fn writeHeadNumberDelegate) {
	m.writeHeadNumberFn = fn
}

func (m *MockStorage) WriteForks(forks []types.Hash) error {
	if m.writeForksFn != nil {
		return m.writeForksFn(forks)
	}

	return nil
}

func (m *MockStorage) HookWriteForks(fn writeForksDelegate) {
	m.writeForksFn = fn
}

func (m *MockStorage) ReadForks() ([]types.Hash, error) {
	if m.readForksFn != nil {
		return m.readForksFn()
	}

	return []types.Hash{}, nil
}

func (m *MockStorage) HookReadForks(fn readForksDelegate) {
	m.readForksFn = fn
}

func (m *MockStorage) WriteTotalDifficulty(hash types.Hash, diff *big.Int) error {
	if m.writeTotalDifficultyFn != nil {
		return m.writeTotalDifficultyFn(hash, diff)
	}

	return nil
}

func (m *MockStorage) HookWriteTotalDifficulty(fn writeTotalDifficultyDelegate) {
	m.writeTotalDifficultyFn = fn
}

func (m *MockStorage) ReadTotalDifficulty(hash types.Hash) (*big.Int, bool) {
	if m.readTotalDifficultyFn != nil {
		return m.readTotalDifficultyFn(hash)
	}

	return big.NewInt(0), true
}

func (m *MockStorage) HookReadTotalDifficulty(fn readTotalDifficultyDelegate) {
	m.readTotalDifficultyFn = fn
}

func (m *MockStorage) WriteHeader(h *types.Header) error {
	if m.writeHeaderFn != nil {
		return m.writeHeaderFn(h)
	}

	return nil
}

func (m *MockStorage) HookWriteHeader(fn writeHeaderDelegate) {
	m.writeHeaderFn = fn
}

func (m *MockStorage) ReadHeader(hash types.Hash) (*types.Header, error) {
	if m.readHeaderFn != nil {
		return m.readHeaderFn(hash)
	}

	return &types.Header{}, nil
}

func (m *MockStorage) HookReadHeader(fn readHeaderDelegate) {
	m.readHeaderFn = fn
}

func (m *MockStorage) WriteCanonicalHeader(h *types.Header, diff *big.Int) error {
	if m.writeCanonicalHeaderFn != nil {
		return m.writeCanonicalHeaderFn(h, diff)
	}

	return nil
}

func (m *MockStorage) HookWriteCanonicalHeader(fn writeCanonicalHeaderDelegate) {
	m.writeCanonicalHeaderFn = fn
}

func (m *MockStorage) WriteBody(hash types.Hash, body *types.Body) error {
	if m.writeBodyFn != nil {
		return m.writeBodyFn(hash, body)
	}

	return nil
}

func (m *MockStorage) HookWriteBody(fn writeBodyDelegate) {
	m.writeBodyFn = fn
}

func (m *MockStorage) ReadBody(hash types.Hash) (*types.Body, error) {
	if m.readBodyFn != nil {
		return m.readBodyFn(hash)
	}

	return &types.Body{}, nil
}

func (m *MockStorage) HookReadBody(fn readBodyDelegate) {
	m.readBodyFn = fn
}

func (m *MockStorage) WriteReceipts(hash types.Hash, receipts []*types.Receipt) error {
	if m.writeReceiptsFn != nil {
		return m.writeReceiptsFn(hash, receipts)
	}

	return nil
}

func (m *MockStorage) HookWriteReceipts(fn writeReceiptsDelegate) {
	m.writeReceiptsFn = fn
}

func (m *MockStorage) ReadReceipts(hash types.Hash) ([]*types.Receipt, error) {
	if m.readReceiptsFn != nil {
		return m.readReceiptsFn(hash)
	}

	return []*types.Receipt{}, nil
}

func (m *MockStorage) HookReadReceipts(fn readReceiptsDelegate) {
	m.readReceiptsFn = fn
}

func (m *MockStorage) WriteTxLookup(hash types.Hash, blockHash types.Hash) error {
	if m.writeTxLookupFn != nil {
		return m.writeTxLookupFn(hash, blockHash)
	}

	return nil
}

func (m *MockStorage) HookWriteTxLookup(fn writeTxLookupDelegate) {
	m.writeTxLookupFn = fn
}

func (m *MockStorage) ReadTxLookup(hash types.Hash) (types.Hash, bool) {
	if m.readTxLookupFn != nil {
		return m.readTxLookupFn(hash)
	}

	return types.Hash{}, true
}

func (m *MockStorage) HookReadTxLookup(fn readTxLookupDelegate) {
	m.readTxLookupFn = fn
}

func (m *MockStorage) Close() error {
	if m.closeFn != nil {
		return m.closeFn()
	}

	return nil
}

func (m *MockStorage) HookClose(fn closeDelegate) {
	m.closeFn = fn
}
