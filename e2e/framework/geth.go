package framework

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/apex-fusion/nexus/helper/tests"
	"github.com/umbracle/ethgo"
	"os"
	"os/exec"
	"strings"
)

func (t *TestServer) initGeth() error {
	args := []string{
		fmt.Sprintf("--datadir=%s", t.Config.GethDataDir),
		"init",
		fmt.Sprintf("%s/%s", t.Config.RootDir, "geth-genesis.json"),
	}

	cmd := exec.Command(t.resolveGethBinary(), args...) //nolint:gosec
	cmd.Dir = t.Config.RootDir

	_, err := cmd.Output()

	if err != nil {
		return err
	}

	t.Config.JWTHex, err = generateJWTSecret(fmt.Sprintf("%s/%s/jwt.hex", t.Config.RootDir, t.Config.GethDataDir))
	if err != nil {
		return err
	}

	return nil
}

// startGeth starts go-ethereum and if the bootnode is set, it will also fetch the enode.
func (t *TestServer) startGeth(ctx context.Context, bootnodeEnode string) error {
	args := []string{
		"--verbosity=4",
		"--http",
		"--http.api=eth,net,web3,admin,txpool",
		"--http.addr=0.0.0.0",
		"--http.corsdomain=*",
		fmt.Sprintf("--http.port=%d", t.Config.JSONRPCPort),
		"--authrpc.vhosts=*",
		"--authrpc.addr=0.0.0.0",
		fmt.Sprintf("--authrpc.jwtsecret=%s/jwt.hex", t.Config.GethDataDir),
		fmt.Sprintf("--authrpc.port=%d", t.Config.EnginePort),
		"--datadir=/geth",
		"--syncmode=full",
		"--state.scheme=path", // in order to use pebbledb
		fmt.Sprintf("--datadir=%s", t.Config.GethDataDir),
		fmt.Sprintf("--port=%d", t.Config.DevP2PPort),
		"--nat=extip:127.0.0.1",
		"--netrestrict=127.0.0.1/32",
	}

	// If bootnodeEnode arg is passed, then we set it
	if bootnodeEnode != "" {
		args = append(args, fmt.Sprintf("--bootnodes=%s", strings.TrimSuffix(bootnodeEnode, "?discport=0")))
	}

	t.ReleaseReservedPorts()

	// Start the Geth server
	t.gethCmd = exec.Command(t.resolveGethBinary(), args...)
	t.gethCmd.Dir = t.Config.RootDir

	stdout := t.GetStdout("geth")
	t.gethCmd.Stdout = stdout
	t.gethCmd.Stderr = stdout

	if err := t.gethCmd.Start(); err != nil {
		return err
	}

	block, err := tests.RetryUntilTimeout(ctx, func() (interface{}, bool) {
		block, err := t.JSONRPC().Eth().GetBlockByNumber(0, false)

		if err != nil {
			return nil, true
		}

		return block, false
	})
	if err != nil {
		return fmt.Errorf("failed to fetch genesis block: %w", err)
	}

	castBlock := block.(*ethgo.Block)
	t.Config.ExecutionGenesisBlockHash = castBlock.Hash.String()

	// If bootnodeEnode arg is not passed, then we just need to fetch it
	if bootnodeEnode == "" {
		var result map[string]interface{}
		err := t.JSONRPC().Call("admin_nodeInfo", &result)
		if err != nil {
			return fmt.Errorf("failed to fetch enode for bootnodeEnode: %w", err)
		}

		enode, ok := result["enode"].(string)
		if !ok {
			return fmt.Errorf("failed to parse enode from response: %w", err)
		}

		t.Config.GethBootnodeEnode = enode
	}

	return nil
}

func (t *TestServer) resolveGethBinary() string {
	bin := os.Getenv("GETH_BINARY")
	if bin != "" {
		return bin
	}
	// fallback
	return fmt.Sprintf("%s/%s", t.Config.RootDir, "nexus-geth")
}

// GenerateJWTSecret generates a 256-bit random JWT secret and writes it to a file in hexadecimal format.
func generateJWTSecret(filePath string) (string, error) {
	// Create a 32-byte array (256 bits)
	secret := make([]byte, 32)
	_, err := rand.Read(secret)
	if err != nil {
		return "", fmt.Errorf("failed to generate random secret: %w", err)
	}

	// Encode the secret to hexadecimal
	hexSecret := hex.EncodeToString(secret)

	// Write the secret to the given file path
	err = os.WriteFile(filePath, []byte(hexSecret), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write JWT secret to file: %w", err)
	}

	return hexSecret, nil
}
