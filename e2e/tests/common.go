package tests

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"os/exec"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	gTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/jsonrpc"
	"github.com/umbracle/ethgo/wallet"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
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
		"geth-1", "gethsecond-1", "geththird-1", "gethfourth-1", 
		"nexus-1", "nexussecond-1", "nexusthird-1, nexusfourth-1",
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

func getSendTxRawBytes(masterPrivKey string, rpcUrl string, addr string, valueInt int64) (string) {

	client, err := ethclient.Dial(rpcUrl)
    privateKey, err := crypto.HexToECDSA(masterPrivKey)

    publicKey := privateKey.Public()
    publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)

    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        log.Fatal(err)
    }

    value := big.NewInt(valueInt) 
    gasLimit := uint64(21000) // in units
    gasPrice, err := client.SuggestGasPrice(context.Background())

    toAddress := common.HexToAddress(addr)
    var data []byte
    tx := gTypes.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)

    chainID, err := client.NetworkID(context.Background())

    signedTx, err := gTypes.SignTx(tx, gTypes.NewEIP155Signer(chainID), privateKey)
    
    ts := gTypes.Transactions{signedTx}
	b := new(bytes.Buffer)
	ts.EncodeIndex(0, b)
	rawTxBytes := b.Bytes()
    rawTxHex := hex.EncodeToString(rawTxBytes)

    return rawTxHex
}


const (
	DOCKER_ENV_SINGLE = iota
	DOCKER_ENV_MULTI
)

func defaultDelay() {
	time.Sleep(time.Duration(20) * time.Second)
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

func testFetchAndCheckMetaFields (vId string, t  *testing.T, clt *jsonrpc.Client) () {

	t.Run("fetchMetaFields(validator=" + vId+ ")", func(t *testing.T) {
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

func testBalanceGreaterThanZero (vId string, t  *testing.T, clt *jsonrpc.Client, account *wallet.Key) () {
	t.Run("eth_getBalance(validator=" + vId+ ")", func(t *testing.T) {
		
		t.Log("fetching masterAcc's balance...")
		balance, err := clt.Eth().GetBalance(account.Address(), ethgo.Latest)
		require.NoError(t, err)
		require.Less(t, uint64(0), balance.Uint64())
	})
}

func testBlockAreBeingProduced (vId string, t  *testing.T, clt *jsonrpc.Client) {

	currentBlockNumber := uint64(0)
	t.Run("blocks being produced (validator=" + vId+ ")", func(t *testing.T) {
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

func testBroadcastTx (vId string, t  *testing.T, clt *jsonrpc.Client, masterAccPrivKey string, rpcUrl string) {
	t.Run("sendTransaction", func(t *testing.T) {

		recipientAddr := "0x4592d8f8d7b001e72cb26a73e4fa1806a51ac79d"
		value := int64(10000)

		privateKey, _ := hex.DecodeString(masterAccPrivKey) 
		wk, _ := wallet.NewWalletFromPrivKey(privateKey)
		
		//record the sender's nonce and reciever's balance before the transaction
		previousSenderNonce, err := clt.Eth().GetNonce(wk.Address(), ethgo.Latest)
		require.NoError(t, err)

		previousReceiverBalance, err := clt.Eth().GetBalance(ethgo.HexToAddress(recipientAddr), ethgo.Latest)
		require.NoError(t, err)
		expectedReceiverBalance := previousReceiverBalance.Int64() + value

		t.Log("sending a transaction to a new account...")
		rawTx := getSendTxRawBytes(masterAccPrivKey, rpcUrl, recipientAddr, value)
		rawTxBytes, err := hex.DecodeString(rawTx)
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
		require.Equal(t, previousSenderNonce + 1, actualNonce)

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