package framework

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/apex-fusion/nexus/helper/tests"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/wallet"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
)

func (t *TestServer) Txn(key *wallet.Key) *Txn {
	tt := &Txn{
		key:     key,
		client:  t.JSONRPC().Eth(),
		chainID: t.chainID,
		raw:     &ethgo.Transaction{},
	}

	return tt
}

func (t *TestServer) WaitForReceipt(ctx context.Context, hash ethgo.Hash) (*ethgo.Receipt, error) {
	client := t.JSONRPC()

	type result struct {
		receipt *ethgo.Receipt
		err     error
	}

	res, err := tests.RetryUntilTimeout(ctx, func() (interface{}, bool) {
		receipt, err := client.Eth().GetTransactionReceipt(hash)
		if err != nil && err.Error() != "not found" {
			return result{receipt, err}, false
		}
		if receipt != nil {
			return result{receipt, nil}, false
		}

		return nil, true
	})
	if err != nil {
		return nil, err
	}

	data, ok := res.(result)
	if !ok {
		return nil, errors.New("invalid type assertion")
	}

	return data.receipt, data.err
}

// GetGasTotal waits for the total gas used sum for the passed in
// transactions
func (t *TestServer) GetGasTotal(txHashes []ethgo.Hash) uint64 {
	t.t.Helper()

	var (
		totalGasUsed    = uint64(0)
		receiptErrs     = make([]error, 0)
		receiptErrsLock sync.Mutex
		wg              sync.WaitGroup
	)

	appendReceiptErr := func(receiptErr error) {
		receiptErrsLock.Lock()
		defer receiptErrsLock.Unlock()

		receiptErrs = append(receiptErrs, receiptErr)
	}

	for _, txHash := range txHashes {
		wg.Add(1)

		go func(txHash ethgo.Hash) {
			defer wg.Done()

			ctx, cancelFn := context.WithTimeout(context.Background(), DefaultTimeout)
			defer cancelFn()

			receipt, receiptErr := tests.WaitForReceipt(ctx, t.JSONRPC().Eth(), txHash)
			if receiptErr != nil {
				appendReceiptErr(fmt.Errorf("unable to wait for receipt, %w", receiptErr))

				return
			}

			atomic.AddUint64(&totalGasUsed, receipt.GasUsed)
		}(txHash)
	}

	wg.Wait()

	if len(receiptErrs) > 0 {
		t.t.Fatalf("unable to wait for receipts, %v", receiptErrs)
	}

	return totalGasUsed
}

func (t *TestServer) WaitForReady(ctx context.Context) error {
	_, err := tests.RetryUntilTimeout(ctx, func() (interface{}, bool) {
		num, err := t.GetLatestBlockHeight()

		if num < 1 || err != nil {
			return nil, true
		}

		return nil, false
	})

	return err
}

func (t *TestServer) CallJSONRPC(req map[string]interface{}) map[string]interface{} {
	reqJSON, err := json.Marshal(req)
	if err != nil {
		t.t.Fatal(err)

		return nil
	}

	url := fmt.Sprintf("http://%s", t.JSONRPCAddr())

	//nolint:gosec // this is not used because it can't be defined as a global variable
	response, err := http.Post(url, "application/json", bytes.NewReader(reqJSON))
	if err != nil {
		t.t.Fatalf("failed to send request to JSON-RPC server: %v", err)

		return nil
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.t.Fatalf("JSON-RPC doesn't return ok: %s", response.Status)

		return nil
	}

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		t.t.Fatalf("failed to read HTTP body: %s", err)

		return nil
	}

	result := map[string]interface{}{}

	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		t.t.Fatalf("failed to convert json to object: %s", err)

		return nil
	}

	return result
}

func (t *TestServer) GetLatestBlockHeight() (uint64, error) {
	return t.JSONRPC().Eth().BlockNumber()
}
