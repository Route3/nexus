package framework

import (
	"context"
	"fmt"
	"github.com/apex-fusion/nexus/helper/tests"
	"github.com/apex-fusion/nexus/types"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/jsonrpc"
	"github.com/umbracle/ethgo/wallet"
	"math/big"
	"testing"
)

type PreparedTransaction struct {
	From     types.Address
	GasPrice *big.Int
	Gas      uint64
	To       *types.Address
	Value    *big.Int
	Input    []byte
}

type Transaction struct {
	key     *wallet.Key
	client  *jsonrpc.Eth
	hash    *ethgo.Hash
	receipt *ethgo.Receipt
	raw     *ethgo.Transaction
	chainID *big.Int

	sendErr error
	waitErr error
}

func (t *Transaction) Deploy(input []byte) *Transaction {
	t.raw.Input = input

	return t
}

func (t *Transaction) Transfer(to ethgo.Address, value *big.Int) *Transaction {
	t.raw.To = &to
	t.raw.Value = value

	return t
}

func (t *Transaction) Value(value *big.Int) *Transaction {
	t.raw.Value = value

	return t
}

func (t *Transaction) To(to ethgo.Address) *Transaction {
	t.raw.To = &to

	return t
}

func (t *Transaction) GasLimit(gas uint64) *Transaction {
	t.raw.Gas = gas

	return t
}

func (t *Transaction) GasPrice(price uint64) *Transaction {
	t.raw.GasPrice = price

	return t
}

func (t *Transaction) Nonce(nonce uint64) *Transaction {
	t.raw.Nonce = nonce

	return t
}

func (t *Transaction) sendImpl() error {
	// populate default values
	t.raw.Gas = 1048576
	t.raw.GasPrice = 1048576

	if t.raw.Nonce == 0 {
		nextNonce, err := t.client.GetNonce(t.key.Address(), ethgo.Latest)
		if err != nil {
			return fmt.Errorf("failed to get nonce: %w", err)
		}

		t.raw.Nonce = nextNonce
	}

	signer := wallet.NewEIP155Signer(t.chainID.Uint64())

	signedTxn, err := signer.SignTx(t.raw, t.key)
	if err != nil {
		return err
	}

	data, _ := signedTxn.MarshalRLPTo(nil)

	txHash, err := t.client.SendRawTransaction(data)
	if err != nil {
		return fmt.Errorf("failed to send transaction: %w", err)
	}

	t.hash = &txHash

	return nil
}

func (t *Transaction) Send() *Transaction {
	if t.hash != nil {
		panic("BUG: txn already sent")
	}

	t.sendErr = t.sendImpl()

	return t
}

func (t *Transaction) Receipt() *ethgo.Receipt {
	return t.receipt
}

//nolint:thelper
func (t *Transaction) NoFail(tt *testing.T) {
	t.Wait()

	if t.sendErr != nil {
		tt.Fatal(t.sendErr)
	}

	if t.waitErr != nil {
		tt.Fatal(t.waitErr)
	}

	if t.receipt.Status != 1 {
		tt.Fatal("txn failed with status 0")
	}
}

func (t *Transaction) Complete() bool {
	if t.sendErr != nil {
		// txn failed during sending
		return true
	}

	if t.waitErr != nil {
		// txn failed during waiting
		return true
	}

	if t.receipt != nil {
		// txn was mined
		return true
	}

	return false
}

func (t *Transaction) Wait() {
	if t.Complete() {
		return
	}

	ctx, cancelFn := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancelFn()

	receipt, err := tests.WaitForReceipt(ctx, t.client, *t.hash)
	if err != nil {
		t.waitErr = err
	} else {
		t.receipt = receipt
	}
}
