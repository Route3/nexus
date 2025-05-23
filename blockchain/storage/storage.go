package storage

import (
	"math/big"

	"github.com/apex-fusion/nexus/types"
	"github.com/hashicorp/go-hclog"
)

// Storage is a generic blockchain storage
type Storage interface {
	ReadCanonicalHash(n uint64) (types.Hash, bool)
	WriteCanonicalHash(n uint64, hash types.Hash) error

	ReadHeadHash() (types.Hash, bool)
	ReadHeadNumber() (uint64, bool)
	WriteHeadHash(h types.Hash) error
	WriteHeadNumber(uint64) error

	WriteForks(forks []types.Hash) error
	ReadForks() ([]types.Hash, error)

	WriteTotalDifficulty(hash types.Hash, diff *big.Int) error
	ReadTotalDifficulty(hash types.Hash) (*big.Int, bool)

	WriteHeader(h *types.Header) error
	ReadHeader(hash types.Hash) (*types.Header, error)

	WriteCanonicalHeader(h *types.Header, diff *big.Int) error

	WriteBody(hash types.Hash, body *types.Body) error
	ReadBody(hash types.Hash) (*types.Body, error)

	Close() error
}

// Factory is a factory method to create a blockchain storage
type Factory func(config map[string]interface{}, logger hclog.Logger) (Storage, error)
