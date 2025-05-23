package tests

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"net"
	"testing"
	"time"

	libp2pCrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/umbracle/ethgo"

	"github.com/apex-fusion/nexus/crypto"
	"github.com/apex-fusion/nexus/types"
	"github.com/stretchr/testify/assert"
	"github.com/umbracle/ethgo/jsonrpc"
)

var (
	ErrTimeout = errors.New("timeout")
)

func GenerateKeyAndAddr(t *testing.T) (*ecdsa.PrivateKey, types.Address) {
	t.Helper()

	key, err := crypto.GenerateECDSAKey()

	assert.NoError(t, err)

	addr := crypto.PubKeyToAddress(&key.PublicKey)

	return key, addr
}

func GenerateTestMultiAddr(t *testing.T) multiaddr.Multiaddr {
	t.Helper()

	priv, _, err := libp2pCrypto.GenerateKeyPair(libp2pCrypto.Secp256k1, 256)
	if err != nil {
		t.Fatalf("Unable to generate key pair, %v", err)
	}

	nodeID, err := peer.IDFromPrivateKey(priv)
	assert.NoError(t, err)

	port, portErr := GetFreePort()
	if portErr != nil {
		t.Fatalf("Unable to fetch free port, %v", portErr)
	}

	addr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d/p2p/%s", port, nodeID))
	assert.NoError(t, err)

	return addr
}

func RetryUntilTimeout(ctx context.Context, f func() (interface{}, bool)) (interface{}, error) {
	type result struct {
		data interface{}
		err  error
	}

	resCh := make(chan result, 1)

	go func() {
		defer close(resCh)

		for {
			select {
			case <-ctx.Done():
				resCh <- result{nil, ErrTimeout}

				return
			default:
				res, retry := f()
				if !retry {
					resCh <- result{res, nil}

					return
				}
			}
			time.Sleep(time.Second)
		}
	}()

	res := <-resCh

	return res.data, res.err
}

func WaitForNonce(
	ctx context.Context,
	ethClient *jsonrpc.Eth,
	addr ethgo.Address,
	expectedNonce uint64,
) (
	interface{},
	error,
) {
	type result struct {
		nonce uint64
		err   error
	}

	resObj, err := RetryUntilTimeout(ctx, func() (interface{}, bool) {
		nonce, err := ethClient.GetNonce(addr, ethgo.Latest)
		if err != nil {
			// error -> stop retrying
			return result{nonce, err}, false
		}

		if nonce >= expectedNonce {
			// match -> return result
			return result{nonce, nil}, false
		}

		// continue retrying
		return nil, true
	})

	if err != nil {
		return nil, err
	}

	res, ok := resObj.(result)
	if !ok {
		return nil, errors.New("invalid type assertion")
	}

	return res.nonce, res.err
}

// WaitForReceipt waits transaction receipt
func WaitForReceipt(ctx context.Context, client *jsonrpc.Eth, hash ethgo.Hash) (*ethgo.Receipt, error) {
	type result struct {
		receipt *ethgo.Receipt
		err     error
	}

	res, err := RetryUntilTimeout(ctx, func() (interface{}, bool) {
		receipt, err := client.GetTransactionReceipt(hash)
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

// GetFreePort asks the kernel for a free open port that is ready to use
func GetFreePort() (port int, err error) {
	var addr *net.TCPAddr

	if addr, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener

		if l, err = net.ListenTCP("tcp", addr); err == nil {
			defer func(l *net.TCPListener) {
				_ = l.Close()
			}(l)

			netAddr, ok := l.Addr().(*net.TCPAddr)
			if !ok {
				return 0, errors.New("invalid type assert to TCPAddr")
			}

			return netAddr.Port, nil
		}
	}

	return
}

type GenerateTxReqParams struct {
	Nonce         uint64
	ReferenceAddr types.Address
	ReferenceKey  *ecdsa.PrivateKey
	ToAddress     types.Address
	GasPrice      *big.Int
	Value         *big.Int
	Input         []byte
}
