package tests

import (
	"context"
	"encoding/hex"
	"fmt"
	//"github.com/docker/docker/client"
	"log"
	"math/big"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/jsonrpc"
	"github.com/umbracle/ethgo/wallet"
	//"github.com/docker/docker/api/types/container"
	//"github.com/docker/docker/api/types/filters"
	//"github.com/docker/docker/client"
)

func basicSingleSetup(t *testing.T) (*jsonrpc.Client, *wallet.Key) {

	startDockerEnv(t, DOCKER_ENV_SINGLE)

	testConfig := LoadSingleTestConfig()

	clt, err := jsonrpc.NewClient(testConfig.rpcUrls[0])
	require.NoError(t, err)

	privateKey, _ := hex.DecodeString(testConfig.masterAccountPrivateKey)
	masterAcc, _ := wallet.NewWalletFromPrivKey(privateKey)

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	require.NoError(t, err)

	timeout := 1 * time.Minute

	containerNames := []string{"geth-1", "nexus-1"}
	err = waitForServices(cli, timeout, t, containerNames)

	return clt, masterAcc
}

func basicMultiSetup(t *testing.T) ([]*jsonrpc.Client, *wallet.Key) {

	startDockerEnv(t, DOCKER_ENV_MULTI)

	testConfig := LoadMultiTestConfig()

	privateKey, _ := hex.DecodeString(testConfig.masterAccountPrivateKey)
	masterAcc, _ := wallet.NewWalletFromPrivKey(privateKey)

	var clts []*jsonrpc.Client

	for i := 0; i < len(testConfig.rpcUrls); i++ {

		clt, err := jsonrpc.NewClient(testConfig.rpcUrls[i])
		require.NoError(t, err)

		clts = append(clts, clt)
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	require.NoError(t, err)

	timeout := 1 * time.Minute
	containerNames := []string{
		"geth-0-1", "geth-1-1", "geth-2-1", "geth-3-1",
		"nexus-0-1", "nexus-1-1", "nexus-2-1, nexus-3-1",
	} // Adjust container names if necessary
	err = waitForServices(cli, timeout, t, containerNames)

	return clts, masterAcc
}

func startDockerEnv(t *testing.T, envType int) (err error) {

	var cmd *exec.Cmd
	if envType == DOCKER_ENV_SINGLE {
		cmd = exec.Command("/bin/sh", "-c", "cd ../.. && make run-single")
	} else if envType == DOCKER_ENV_MULTI {
		cmd = exec.Command("/bin/sh", "-c", "cd ../.. && make run-multi")
	}

	err = cmd.Run()
	require.NoError(t, err)

	// Artificial delay, used to wait for the docker services to start
	time.Sleep(time.Duration(10) * time.Second)

	return err
}

func cleanupDockerEnv(t *testing.T) (err error) {
	cmd := exec.Command("/bin/sh", "-c", "cd ../.. && make clean-single && make clean-multi")

	err = cmd.Run()
	if err != nil {
		t.Errorf("Error cleaning up docker-compose: %v\n", err)
		return
	}

	return
}

func getSendTxRawBytes(masterPrivKey string, nonce uint64, chainId *big.Int, addr string, valueInt int64) []byte {
	privateBytes, _ := hex.DecodeString(masterPrivKey)
	sender, _ := wallet.NewWalletFromPrivKey(privateBytes)

	signer := wallet.NewEIP155Signer(chainId.Uint64())

	transaction := &ethgo.Transaction{
		Value:    big.NewInt(valueInt),
		Nonce:    nonce,
		Gas:      1048576,
		GasPrice: 1048576,
	}

	signedTransaction, err := signer.SignTx(transaction, sender)
	if err != nil {
		log.Fatal(err)
	}

	data, _ := signedTransaction.MarshalRLPTo(nil)
	return data
}

const (
	DOCKER_ENV_SINGLE = iota
	DOCKER_ENV_MULTI
)

func defaultDelay() {
	time.Sleep(time.Duration(45) * time.Second)
}

// waitForServices waits for the Geth nodes to become healthy.
func waitForServices(cli *client.Client, timeout time.Duration, t *testing.T, containerNames []string) error {
	ctx := context.Background()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	timeoutChan := time.After(timeout)

	for {
		select {
		case <-ticker.C:
			filtersArgs := filters.NewArgs()
			filtersArgs.Add("status", "running")
			for _, name := range containerNames {
				filtersArgs.Add("name", name)
			}

			containers, err := cli.ContainerList(ctx, container.ListOptions{
				Filters: filtersArgs,
			})
			if err != nil {
				t.Errorf("Error listing containers: %v", err)
				return err
			}

			if len(containers) != len(containerNames) {
				t.Logf("Expected %d containers, found %d", len(containerNames), len(containers))
				continue
			}

			allHealthy := true
			for _, container := range containers {
				// Inspect the container to get health status
				inspect, err := cli.ContainerInspect(ctx, container.ID)
				if err != nil {
					t.Errorf("Error inspecting container %s: %v", container.ID, err)
					return err
				}

				if inspect.State.Health == nil {
					t.Logf("Container %s has no health status yet", container.Names[0])
					allHealthy = false
					break
				}

				healthStatus := inspect.State.Health.Status
				t.Logf("Container %s health status: %s", container.Names[0], healthStatus)
				if healthStatus != "healthy" {
					allHealthy = false
					break
				}
			}

			if allHealthy {
				t.Log("All specified services are healthy!")
				return nil
			}
		case <-timeoutChan:
			return fmt.Errorf("timeout waiting for services to become healthy")
		}
	}
}

func testFetchAndCheckMetaFields(vId string, t *testing.T, clt *jsonrpc.Client) {

	t.Run("fetchMetaFields(validator="+vId+")", func(t *testing.T) {
		t.Log("fetching chain's id...")
		_, err := clt.Eth().ChainID()
		require.NoError(t, err)

		t.Log("fetching chain's gas price...")
		gasPrice, err := clt.Eth().GasPrice()
		require.NoError(t, err)
		require.Greater(t, gasPrice, uint64(0))

		t.Log("fetching chain's block number...")
		blockNumber, err := clt.Eth().BlockNumber()
		require.NoError(t, err)
		require.Less(t, uint64(0), blockNumber)
		currentBlockNumber := blockNumber

		t.Log("fetching chain's fee history...")
		feeHistory, err := clt.Eth().FeeHistory(ethgo.BlockNumber(currentBlockNumber), ethgo.BlockNumber(currentBlockNumber))
		require.NoError(t, err)
		require.Less(t, uint64(0), feeHistory.BaseFee[0].Uint64())
	})
}

func testBalanceGreaterThanZero(vId string, t *testing.T, clt *jsonrpc.Client, account *wallet.Key) {
	t.Run("eth_getBalance(validator="+vId+")", func(t *testing.T) {

		t.Log("fetching masterAcc's balance...")
		balance, err := clt.Eth().GetBalance(account.Address(), ethgo.Latest)
		require.NoError(t, err)
		require.Less(t, uint64(0), balance.Uint64())
	})
}

func testBlockAreBeingProduced(vId string, t *testing.T, clt *jsonrpc.Client) {

	currentBlockNumber := uint64(0)
	t.Run("blocks being produced (validator="+vId+")", func(t *testing.T) {
		nRepeats := 3
		for i := 0; i < nRepeats; i++ {
			t.Log("fetching current block number...")
			defaultDelay()
			blockNumber, err := clt.Eth().BlockNumber()
			require.NoError(t, err)
			require.Less(t, currentBlockNumber, blockNumber)
			currentBlockNumber = blockNumber
		}
	})
}

func testBroadcastTx(vId string, t *testing.T, clt *jsonrpc.Client, masterAccPrivKey string, rpcUrl string) {
	t.Skip()
	t.Run("sendTransaction", func(t *testing.T) {

		t.Log("Running testBroadcastTx... for validator:", vId)

		recipientAddr := "0x4592d8f8d7b001e72cb26a73e4fa1806a51ac79d"
		value := int64(10000)

		privateKey, err := hex.DecodeString(masterAccPrivKey)
		require.NoError(t, err)

		wk, err := wallet.NewWalletFromPrivKey(privateKey)
		require.NoError(t, err)

		//record the sender's nonce and reciever's balance before the transaction
		previousSenderNonce, err := clt.Eth().GetNonce(wk.Address(), ethgo.Latest)
		require.NoError(t, err)

		previousReceiverBalance, err := clt.Eth().GetBalance(ethgo.HexToAddress(recipientAddr), ethgo.Latest)
		require.NoError(t, err)
		expectedReceiverBalance := previousReceiverBalance.Int64() + value

		chainId, err := clt.Eth().ChainID()
		require.NoError(t, err)

		t.Log("sending a transaction to a new account...")
		rawTxBytes := getSendTxRawBytes(masterAccPrivKey, previousSenderNonce, chainId, recipientAddr, value)
		require.NoError(t, err)

		txHash, err := clt.Eth().SendRawTransaction(rawTxBytes)
		require.NoError(t, err)

		t.Log("checking tx inclusion... for:", txHash)
		tx, err := clt.Eth().GetTransactionByHash(txHash)
		require.NoError(t, err)
		require.NotEqual(t, nil, tx.BlockNumber)

		t.Log("checking sender's nonce...")
		defaultDelay()
		actualNonce, err := clt.Eth().GetNonce(wk.Address(), ethgo.Latest)
		require.NoError(t, err)
		require.Equal(t, previousSenderNonce+1, actualNonce)

		t.Log("checking receiver's balance...")
		actualReceiverBalance, err := clt.Eth().GetBalance(ethgo.HexToAddress(recipientAddr), ethgo.Latest)
		require.NoError(t, err)
		require.Equal(t, expectedReceiverBalance, actualReceiverBalance.Int64())

		t.Log("checking receiver's nonce...")
		actualNonce, err = clt.Eth().GetNonce(ethgo.HexToAddress(recipientAddr), ethgo.Latest)
		require.NoError(t, err)
		require.Equal(t, uint64(0), actualNonce)
	})
}
