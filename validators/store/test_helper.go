package store

import (
	"github.com/apex-fusion/nexus/consensus/ibft/signer"
	"github.com/apex-fusion/nexus/types"
)

// Utilities for test
const (
	TestEpochSize = 100
)

func NewMockGetSigner(s signer.Signer) func(uint64) (signer.Signer, error) {
	return func(u uint64) (signer.Signer, error) {
		return s, nil
	}
}

type MockBlockchain struct {
	HeaderFn            func() *types.Header
	GetHeaderByNumberFn func(uint64) (*types.Header, bool)
}

func (m *MockBlockchain) Header() *types.Header {
	return m.HeaderFn()
}

func (m *MockBlockchain) GetHeaderByNumber(height uint64) (*types.Header, bool) {
	return m.GetHeaderByNumberFn(height)
}
