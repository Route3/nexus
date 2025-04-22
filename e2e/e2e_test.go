package e2e

import (
	"context"
	"fmt"
	"github.com/apex-fusion/nexus/helper/hex"
	"github.com/apex-fusion/nexus/helper/tests"
	"github.com/stretchr/testify/require"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/wallet"
	"math/big"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/apex-fusion/nexus/e2e/framework"
)

const (
	numNonValidators = 2
	ibftMinNodes     = 4
	numberOfServers  = ibftMinNodes + numNonValidators
	desiredHeight    = 10
)

// TestMain will run once before any tests.
func TestMain(m *testing.M) {
	// Build the Nexus binary with debugging flags enabled
	if err := framework.Build(); err != nil {
		panic("Cannot build Nexus binary: " + err.Error() + "")
	}

	// Run the tests
	m.Run()
}

// TestBlockProduction is a shorter smoke test
func TestBlockProduction(t *testing.T) {
	// Start IBFT cluster (4 Validator + 2 Non-Validator)
	serverManager, err := framework.NewServerManager(
		t,
		numberOfServers,
		func(i int, config *framework.TestServerConfig) {
			if i >= ibftMinNodes {
				// Other nodes should not be in the validator set
				dirPrefix := "nexus-non-validator-"
				config.SetIBFTDir(fmt.Sprintf("%s%d", dirPrefix, i))
			}
		})

	require.NoError(t, err)

	startContext, startCancelFn := context.WithTimeout(context.Background(), time.Minute)
	defer startCancelFn()
	serverManager.StartServers(startContext)

	// All nodes should have mined the same block eventually
	waitErrors := framework.WaitForServersToSeal(serverManager.Servers, desiredHeight)

	if len(waitErrors) != 0 {
		t.Fatalf("Unable to wait for all nodes to seal blocks, %v", waitErrors)
	}

}

// TestE2E is a full test that does transaction propagation etc.
func TestE2E(t *testing.T) {
	// Start IBFT cluster (4 Validator + 2 Non-Validator)
	serverManager, err := framework.NewServerManager(
		t,
		numberOfServers,
		func(i int, config *framework.TestServerConfig) {
			if i >= ibftMinNodes {
				// Other nodes should not be in the validator set
				dirPrefix := "nexus-non-validator-"
				config.SetIBFTDir(fmt.Sprintf("%s%d", dirPrefix, i))
			}
		})

	require.NoError(t, err)

	startContext, startCancelFn := context.WithTimeout(context.Background(), time.Minute)
	defer startCancelFn()
	serverManager.StartServers(startContext)

	recipientWallet, err := wallet.GenerateKey()
	require.NoError(t, err)

	preminedWallet, err := wallet.NewWalletFromPrivKey(hex.MustDecodeHex(framework.PREMINE_PRIVATE_KEY))
	require.NoError(t, err)

	expectedPreminedBalance := hex.DecodeHexToBig(strings.TrimPrefix(framework.PREMINE_BALANCE, "0x"))
	halfPremineBalance := big.NewInt(0).Div(expectedPreminedBalance, big.NewInt(2))

	for i, s := range serverManager.Servers {
		client := s.JSONRPC().Eth()

		preminedBalance, err := client.GetBalance(preminedWallet.Address(), ethgo.Latest)
		require.NoError(t, err, fmt.Sprintf("Server %d should be able to fetch balance of address", i))
		require.Equal(t, expectedPreminedBalance, preminedBalance, fmt.Sprintf("Server %d should have exact premined balance", i))
	}

	for i, s := range serverManager.Servers {
		client := s.JSONRPC().Eth()

		recipientBalance, err := client.GetBalance(recipientWallet.Address(), ethgo.Latest)
		require.NoError(t, err, fmt.Sprintf("Server %d should be able to fetch balance of receiving address before sending", i))
		require.Equal(t, recipientBalance, big.NewInt(0), fmt.Sprintf("Server %d should have 0 recipient balance before sending", i))
	}

	// Send transaction from first server
	transaction := serverManager.Servers[0].Transaction(preminedWallet).Transfer(recipientWallet.Address(), halfPremineBalance).Send()
	transaction.NoFail(t)

	// Wait for receipts on all servers
	var wg sync.WaitGroup
	wg.Add(numberOfServers)

	for i, s := range serverManager.Servers {
		go func(i int, s *framework.TestServer) {
			defer wg.Done()
			ctx, cancelFn := context.WithTimeout(context.Background(), framework.DefaultTimeout)
			defer cancelFn()

			_, receiptErr := tests.WaitForReceipt(ctx, s.JSONRPC().Eth(), *transaction.Hash)
			require.NoError(t, receiptErr, fmt.Sprintf("Server %d should be able to fetch receipt", i))
		}(i, s)
	}
	wg.Wait()

	// Check balances on all servers
	for i, s := range serverManager.Servers {
		client := s.JSONRPC().Eth()

		balance, err := client.GetBalance(recipientWallet.Address(), ethgo.Latest)
		require.NoError(t, err, fmt.Sprintf("Server %d should be able to fetch balance of receiving address after sending", i))
		require.Equal(t, halfPremineBalance, balance, fmt.Sprintf("Server #%d should have exact balance on recipient", i))
	}

	// TODO: Add tests:
	// 			- for checking validator fee address balances after paid gas fees
}
