package jsonrpc

import (
	"errors"
	"fmt"

	"github.com/apex-fusion/nexus/types"
)

var (
	ErrHeaderNotFound           = errors.New("header not found")
	ErrLatestNotFound           = errors.New("latest header not found")
	ErrNegativeBlockNumber      = errors.New("invalid argument 0: block number must not be negative")
	ErrFailedFetchGenesis       = errors.New("error fetching genesis block header")
	ErrNoDataInContractCreation = errors.New("contract creation without data provided")
)

type latestHeaderGetter interface {
	Header() *types.Header
}

// GetNumericBlockNumber returns block number based on current state or specified number
func GetNumericBlockNumber(number BlockNumber, store latestHeaderGetter) (uint64, error) {
	switch number {
	case LatestBlockNumber, PendingBlockNumber:
		latest := store.Header()
		if latest == nil {
			return 0, ErrLatestNotFound
		}

		return latest.Number, nil

	case EarliestBlockNumber:
		return 0, nil

	default:
		if number < 0 {
			return 0, ErrNegativeBlockNumber
		}

		return uint64(number), nil
	}
}

type headerGetter interface {
	Header() *types.Header
	GetHeaderByNumber(uint64) (*types.Header, bool)
}

// GetBlockHeader returns a header using the provided number
func GetBlockHeader(number BlockNumber, store headerGetter) (*types.Header, error) {
	switch number {
	case PendingBlockNumber, LatestBlockNumber:
		return store.Header(), nil

	case EarliestBlockNumber:
		header, ok := store.GetHeaderByNumber(uint64(0))
		if !ok {
			return nil, ErrFailedFetchGenesis
		}

		return header, nil

	default:
		// Convert the block number from hex to uint64
		header, ok := store.GetHeaderByNumber(uint64(number))
		if !ok {
			return nil, fmt.Errorf("error fetching block number %d header", uint64(number))
		}

		return header, nil
	}
}

type txLookupAndBlockGetter interface {
	ReadTxLookup(types.Hash) (types.Hash, bool)
	GetBlockByHash(types.Hash, bool) (*types.Block, bool)
}

type blockGetter interface {
	Header() *types.Header
	GetHeaderByNumber(uint64) (*types.Header, bool)
	GetBlockByHash(types.Hash, bool) (*types.Block, bool)
}

func GetHeaderFromBlockNumberOrHash(bnh BlockNumberOrHash, store blockGetter) (*types.Header, error) {
	// The filter is empty, use the latest block by default
	if bnh.BlockNumber == nil && bnh.BlockHash == nil {
		bnh.BlockNumber, _ = createBlockNumberPointer(latest)
	}

	if bnh.BlockNumber != nil {
		// block number
		header, err := GetBlockHeader(*bnh.BlockNumber, store)
		if err != nil {
			return nil, fmt.Errorf("failed to get the header of block %d: %w", *bnh.BlockNumber, err)
		}

		return header, nil
	}

	// block hash
	block, ok := store.GetBlockByHash(*bnh.BlockHash, false)
	if !ok {
		return nil, fmt.Errorf("could not find block referenced by the hash %s", bnh.BlockHash.String())
	}

	return block.Header, nil
}

type nonceGetter interface {
	Header() *types.Header
	GetHeaderByNumber(uint64) (*types.Header, bool)
	GetNonce(types.Address) uint64
	GetAccount(root types.Hash, addr types.Address) error
}

func GetNextNonce(address types.Address, number BlockNumber, store nonceGetter) (uint64, error) {
	if number == PendingBlockNumber {
		// Grab the latest pending nonce from the TxPool
		// If the account is not initialized in the local TxPool,
		// return the latest nonce from the world state
		return store.GetNonce(address), nil
	}

	_, err := GetBlockHeader(number, store)
	if err != nil {
		return 0, err
	}

	return 0, err
}
