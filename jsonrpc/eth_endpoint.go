package jsonrpc

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/hashicorp/go-hclog"

	"github.com/apex-fusion/nexus/chain"
	"github.com/apex-fusion/nexus/helper/common"
	"github.com/apex-fusion/nexus/helper/progress"
	"github.com/apex-fusion/nexus/state"
	"github.com/apex-fusion/nexus/state/runtime"
	"github.com/apex-fusion/nexus/types"
)

type ethTxPoolStore interface {
	// AddTx adds a new transaction to the tx pool
	AddTx(tx *types.Transaction) error

	// GetPendingTx gets the pending transaction from the transaction pool, if it's present
	GetPendingTx(txHash types.Hash) (*types.Transaction, bool)

	// GetNonce returns the next nonce for this address
	GetNonce(addr types.Address) uint64
}

type Account struct {
	Balance *big.Int
	Nonce   uint64
}

type ethStateStore interface {
	GetAccount(root types.Hash, addr types.Address) (*Account, error)
	GetStorage(root types.Hash, addr types.Address, slot types.Hash) ([]byte, error)
	GetForksInTime(blockNumber uint64) chain.ForksInTime
	GetCode(root types.Hash, addr types.Address) ([]byte, error)
}

type ethBlockchainStore interface {
	// Header returns the current header of the chain (genesis if empty)
	Header() *types.Header

	// GetHeaderByNumber gets a header using the provided number
	GetHeaderByNumber(uint64) (*types.Header, bool)

	// GetBlockByHash gets a block using the provided hash
	GetBlockByHash(hash types.Hash, full bool) (*types.Block, bool)

	// GetBlockByNumber returns a block using the provided number
	GetBlockByNumber(num uint64, full bool) (*types.Block, bool)

	// ReadTxLookup returns a block hash in which a given txn was mined
	ReadTxLookup(txnHash types.Hash) (types.Hash, bool)

	// GetReceiptsByHash returns the receipts for a block hash
	GetReceiptsByHash(hash types.Hash) ([]*types.Receipt, error)

	// GetAvgGasPrice returns the average gas price
	GetAvgGasPrice() *big.Int

	// ApplyTxn applies a transaction object to the blockchain
	ApplyTxn(header *types.Header, txn *types.Transaction) (*runtime.ExecutionResult, error)

	// GetSyncProgression retrieves the current sync progression, if any
	GetSyncProgression() *progress.Progression
}

// ethStore provides access to the methods needed by eth endpoint
type ethStore interface {
	ethTxPoolStore
	ethStateStore
	ethBlockchainStore
}

// Eth is the eth jsonrpc endpoint
type Eth struct {
	logger        hclog.Logger
	store         ethStore
	chainID       uint64
	filterManager *FilterManager
	priceLimit    uint64
}

var (
	ErrInsufficientFunds = errors.New("insufficient funds for execution")
)

// ChainId returns the chain id of the client
//
//nolint:stylecheck
func (e *Eth) ChainId() (interface{}, error) {
	return argUintPtr(e.chainID), nil
}

func (e *Eth) Syncing() (interface{}, error) {
	if syncProgression := e.store.GetSyncProgression(); syncProgression != nil {
		// Node is bulk syncing, return the status
		return progression{
			Type:          string(syncProgression.SyncType),
			StartingBlock: argUint64(syncProgression.StartingBlock),
			CurrentBlock:  argUint64(syncProgression.CurrentBlock),
			HighestBlock:  argUint64(syncProgression.HighestBlock),
		}, nil
	}

	// Node is not bulk syncing
	return false, nil
}

// GetBlockByNumber returns information about a block by block number
func (e *Eth) GetBlockByNumber(number BlockNumber, fullTx bool) (interface{}, error) {
	num, err := GetNumericBlockNumber(number, e.store)
	if err != nil {
		return nil, err
	}

	block, ok := e.store.GetBlockByNumber(num, true)
	if !ok {
		return nil, nil
	}

	return toBlock(block, fullTx), nil
}

// GetBlockByHash returns information about a block by hash
func (e *Eth) GetBlockByHash(hash types.Hash, fullTx bool) (interface{}, error) {
	block, ok := e.store.GetBlockByHash(hash, true)
	if !ok {
		return nil, nil
	}

	return toBlock(block, fullTx), nil
}

func (e *Eth) GetBlockTransactionCountByNumber(number BlockNumber) (interface{}, error) {

	return  0, nil
}

// BlockNumber returns current block number
func (e *Eth) BlockNumber() (interface{}, error) {
	h := e.store.Header()
	if h == nil {
		return nil, fmt.Errorf("header has a nil value")
	}

	return argUintPtr(h.Number), nil
}

// SendRawTransaction sends a raw transaction
func (e *Eth) SendRawTransaction(buf argBytes) (interface{}, error) {
	tx := &types.Transaction{}
	if err := tx.UnmarshalRLP(buf); err != nil {
		return nil, err
	}

	tx.ComputeHash()

	if err := e.store.AddTx(tx); err != nil {
		return nil, err
	}

	return tx.Hash.String(), nil
}

// SendTransaction rejects eth_sendTransaction json-rpc call as we don't support wallet management
func (e *Eth) SendTransaction(_ *txnArgs) (interface{}, error) {
	return nil, fmt.Errorf("request calls to eth_sendTransaction method are not supported," +
		" use eth_sendRawTransaction insead")
}

// GetTransactionByHash returns a transaction by its hash.
// If the transaction is still pending -> return the txn with some fields omitted
// If the transaction is sealed into a block -> return the whole txn with all fields
func (e *Eth) GetTransactionByHash(hash types.Hash) (interface{}, error) {
	// findSealedTx is a helper method for checking the world state
	// for the transaction with the provided hash
	findSealedTx := func() *transaction {
		// Check the chain state for the transaction

		return nil
	}

	// findPendingTx is a helper method for checking the TxPool
	// for the pending transaction with the provided hash
	findPendingTx := func() *transaction {
		// Check the TxPool for the transaction if it's pending
		if pendingTx, pendingFound := e.store.GetPendingTx(hash); pendingFound {
			return toPendingTransaction(pendingTx)
		}

		// Transaction not found in the TxPool
		return nil
	}

	// 1. Check the chain state for the txn
	if resultTxn := findSealedTx(); resultTxn != nil {
		return resultTxn, nil
	}

	// 2. Check the TxPool for the txn
	if resultTxn := findPendingTx(); resultTxn != nil {
		return resultTxn, nil
	}

	// Transaction not found in state or TxPool
	e.logger.Warn(
		fmt.Sprintf("Transaction with hash [%s] not found", hash),
	)

	return nil, nil
}

// GetTransactionReceipt returns a transaction receipt by his hash
func (e *Eth) GetTransactionReceipt(hash types.Hash) (interface{}, error) {
	return nil, nil
}

// GetStorageAt returns the contract storage at the index position
func (e *Eth) GetStorageAt(
	address types.Address,
	index types.Hash,
	filter BlockNumberOrHash,
) (interface{}, error) {
	header, err := GetHeaderFromBlockNumberOrHash(filter, e.store)
	if err != nil {
		return nil, err
	}

	// Get the storage for the passed in location
	result, err := e.store.GetStorage(header.StateRoot, address, index)
	if err != nil {
		if errors.Is(err, ErrStateNotFound) {
			return argBytesPtr(types.ZeroHash[:]), nil
		}

		return nil, err
	}

	return argBytesPtr(result), nil
}

// GasPrice returns the average gas price based on the last x blocks
// taking into consideration operator defined price limit
func (e *Eth) GasPrice() (interface{}, error) {
	// Fetch average gas price in uint64
	avgGasPrice := e.store.GetAvgGasPrice().Uint64()

	// Return --price-limit flag defined value if it is greater than avgGasPrice
	return argUint64(common.Max(e.priceLimit, avgGasPrice)), nil
}

// Call executes a smart contract call using the transaction object data
func (e *Eth) Call(arg *txnArgs, filter BlockNumberOrHash) (interface{}, error) {
	header, err := GetHeaderFromBlockNumberOrHash(filter, e.store)
	if err != nil {
		return nil, err
	}

	transaction, err := DecodeTxn(arg, e.store)
	if err != nil {
		return nil, err
	}
	// If the caller didn't supply the gas limit in the message, then we set it to maximum possible => block gas limit
	if transaction.Gas == 0 {
		transaction.Gas = header.GasLimit
	}

	// The return value of the execution is saved in the transition (returnValue field)
	result, err := e.store.ApplyTxn(header, transaction)
	if err != nil {
		return nil, err
	}

	// Check if an EVM revert happened
	if result.Reverted() {
		return nil, constructErrorFromRevert(result)
	}

	if result.Failed() {
		return nil, fmt.Errorf("unable to execute call: %w", result.Err)
	}

	return argBytesPtr(result.ReturnValue), nil
}

// EstimateGas estimates the gas needed to execute a transaction
func (e *Eth) EstimateGas(arg *txnArgs, rawNum *BlockNumber) (interface{}, error) {
	transaction, err := DecodeTxn(arg, e.store)
	if err != nil {
		return nil, err
	}

	number := LatestBlockNumber
	if rawNum != nil {
		number = *rawNum
	}

	// Fetch the requested header
	header, err := GetBlockHeader(number, e.store)
	if err != nil {
		return nil, err
	}

	forksInTime := e.store.GetForksInTime(uint64(number))

	var standardGas uint64
	if transaction.IsContractCreation() && forksInTime.Homestead {
		standardGas = state.TxGasContractCreation
	} else {
		standardGas = state.TxGas
	}

	var (
		lowEnd  = standardGas
		highEnd uint64
	)

	// If the gas limit was passed in, use it as a ceiling
	if transaction.Gas != 0 && transaction.Gas >= standardGas {
		highEnd = transaction.Gas
	} else {
		// If not, use the referenced block number
		highEnd = header.GasLimit
	}

	gasPriceInt := new(big.Int).Set(transaction.GasPrice)
	valueInt := new(big.Int).Set(transaction.Value)

	var availableBalance *big.Int

	// If the sender address is present, figure out how much available funds
	// are we working with
	if transaction.From != types.ZeroAddress {
		// Get the account balance
		// If the account is not initialized yet in state,
		// assume it's an empty account
		accountBalance := big.NewInt(0)
		acc, err := e.store.GetAccount(header.StateRoot, transaction.From)

		if err != nil && !errors.Is(err, ErrStateNotFound) {
			// An unrelated error occurred, return it
			return nil, err
		} else if err == nil {
			// No error when fetching the account,
			// read the balance from state
			accountBalance = acc.Balance
		}

		availableBalance = new(big.Int).Set(accountBalance)

		if transaction.Value != nil {
			if valueInt.Cmp(availableBalance) > 0 {
				return 0, ErrInsufficientFunds
			}

			availableBalance.Sub(availableBalance, valueInt)
		}
	}

	// Recalculate the gas ceiling based on the available funds (if any)
	// and the passed in gas price (if present)
	if gasPriceInt.BitLen() != 0 && // Gas price has been set
		availableBalance != nil && // Available balance is found
		availableBalance.Cmp(big.NewInt(0)) > 0 { // Available balance > 0
		gasAllowance := new(big.Int).Div(availableBalance, gasPriceInt)

		// Check the gas allowance for this account, make sure high end is capped to it
		if gasAllowance.IsUint64() && highEnd > gasAllowance.Uint64() {
			e.logger.Debug(
				fmt.Sprintf(
					"Gas estimation high-end capped by allowance [%d]",
					gasAllowance.Uint64(),
				),
			)

			highEnd = gasAllowance.Uint64()
		}
	}

	// Checks if executor level valid gas errors occurred
	isGasApplyError := func(err error) bool {
		// Not linting this as the underlying error is actually wrapped
		return errors.Is(err, state.ErrNotEnoughIntrinsicGas)
	}

	// Checks if EVM level valid gas errors occurred
	isGasEVMError := func(err error) bool {
		return errors.Is(err, runtime.ErrOutOfGas) ||
			errors.Is(err, runtime.ErrCodeStoreOutOfGas)
	}

	// Checks if the EVM reverted during execution
	isEVMRevertError := func(err error) bool {
		return errors.Is(err, runtime.ErrExecutionReverted)
	}

	// Run the transaction with the specified gas value.
	// Returns a status indicating if the transaction failed and the accompanying error
	testTransaction := func(gas uint64, shouldOmitErr bool) (bool, error) {
		// Create a dummy transaction with the new gas
		txn := transaction.Copy()
		txn.Gas = gas

		result, applyErr := e.store.ApplyTxn(header, txn)

		if applyErr != nil {
			// Check the application error.
			// Gas apply errors are valid, and should be ignored
			if isGasApplyError(applyErr) && shouldOmitErr {
				// Specifying the transaction failed, but not providing an error
				// is an indication that a valid error occurred due to low gas,
				// which will increase the lower bound for the search
				return true, nil
			}

			return true, applyErr
		}

		// Check if an out of gas error happened during EVM execution
		if result.Failed() {
			if isGasEVMError(result.Err) && shouldOmitErr {
				// Specifying the transaction failed, but not providing an error
				// is an indication that a valid error occurred due to low gas,
				// which will increase the lower bound for the search
				return true, nil
			}

			if isEVMRevertError(result.Err) {
				// The EVM reverted during execution, attempt to extract the
				// error message and return it
				return true, constructErrorFromRevert(result)
			}

			return true, result.Err
		}

		return false, nil
	}

	// Start the binary search for the lowest possible gas price
	for lowEnd < highEnd {
		mid := (lowEnd + highEnd) / 2

		failed, testErr := testTransaction(mid, true)
		if testErr != nil &&
			!isEVMRevertError(testErr) {
			// Reverts are ignored in the binary search, but are checked later on
			// during the execution for the optimal gas limit found
			return 0, testErr
		}

		if failed {
			// If the transaction failed => increase the gas
			lowEnd = mid + 1
		} else {
			// If the transaction didn't fail => make this ok value the high end
			highEnd = mid
		}
	}

	// Check if the highEnd is a good value to make the transaction pass
	failed, err := testTransaction(highEnd, false)
	if failed {
		// The transaction shouldn't fail, for whatever reason, at highEnd
		return 0, fmt.Errorf(
			"unable to apply transaction even for the highest gas limit %d: %w",
			highEnd,
			err,
		)
	}

	return argUint64(highEnd), nil
}

// GetFilterLogs returns an array of logs for the specified filter
func (e *Eth) GetFilterLogs(id string) (interface{}, error) {
	logFilter, err := e.filterManager.GetLogFilterFromID(id)
	if err != nil {
		return nil, err
	}

	return e.filterManager.GetLogsForQuery(logFilter.query)
}

// GetLogs returns an array of logs matching the filter options
func (e *Eth) GetLogs(query *LogQuery) (interface{}, error) {
	return e.filterManager.GetLogsForQuery(query)
}

// GetBalance returns the account's balance at the referenced block.
func (e *Eth) GetBalance(address types.Address, filter BlockNumberOrHash) (interface{}, error) {
	header, err := GetHeaderFromBlockNumberOrHash(filter, e.store)
	if err != nil {
		return nil, err
	}

	// Extract the account balance
	acc, err := e.store.GetAccount(header.StateRoot, address)
	if errors.Is(err, ErrStateNotFound) {
		// Account not found, return an empty account
		return argUintPtr(0), nil
	} else if err != nil {
		return nil, err
	}

	return argBigPtr(acc.Balance), nil
}

// GetTransactionCount returns account nonce
func (e *Eth) GetTransactionCount(address types.Address, filter BlockNumberOrHash) (interface{}, error) {
	var (
		blockNumber BlockNumber
		header      *types.Header
		err         error
	)

	// The filter is empty, use the latest block by default
	if filter.BlockNumber == nil && filter.BlockHash == nil {
		filter.BlockNumber, _ = createBlockNumberPointer(latest)
	}

	if filter.BlockNumber == nil {
		header, err = GetHeaderFromBlockNumberOrHash(filter, e.store)
		if err != nil {
			return nil, fmt.Errorf("failed to get header from block hash or block number: %w", err)
		}

		blockNumber = BlockNumber(header.Number)
	} else {
		blockNumber = *filter.BlockNumber
	}

	nonce, err := GetNextNonce(address, blockNumber, e.store)
	if err != nil {
		if errors.Is(err, ErrStateNotFound) {
			return argUintPtr(0), nil
		}

		return nil, err
	}

	return argUintPtr(nonce), nil
}

// GetCode returns account code at given block number
func (e *Eth) GetCode(address types.Address, filter BlockNumberOrHash) (interface{}, error) {
	header, err := GetHeaderFromBlockNumberOrHash(filter, e.store)
	if err != nil {
		return nil, err
	}

	emptySlice := []byte{}
	code, err := e.store.GetCode(header.StateRoot, address)

	if errors.Is(err, ErrStateNotFound) {
		// If the account doesn't exist / is not initialized yet,
		// return the default value
		return "0x", nil
	} else if err != nil {
		return argBytesPtr(emptySlice), err
	}

	return argBytesPtr(code), nil
}

// NewFilter creates a filter object, based on filter options, to notify when the state changes (logs).
func (e *Eth) NewFilter(filter *LogQuery) (interface{}, error) {
	return e.filterManager.NewLogFilter(filter, nil), nil
}

// NewBlockFilter creates a filter in the node, to notify when a new block arrives
func (e *Eth) NewBlockFilter() (interface{}, error) {
	return e.filterManager.NewBlockFilter(nil), nil
}

// GetFilterChanges is a polling method for a filter, which returns an array of logs which occurred since last poll.
func (e *Eth) GetFilterChanges(id string) (interface{}, error) {
	return e.filterManager.GetFilterChanges(id)
}

// UninstallFilter uninstalls a filter with given ID
func (e *Eth) UninstallFilter(id string) (bool, error) {
	return e.filterManager.Uninstall(id), nil
}

// Unsubscribe uninstalls a filter in a websocket
func (e *Eth) Unsubscribe(id string) (bool, error) {
	return e.filterManager.Uninstall(id), nil
}
