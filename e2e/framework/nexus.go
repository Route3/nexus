package framework

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/apex-fusion/nexus/command/genesis"
	"github.com/apex-fusion/nexus/command/server"
	"github.com/apex-fusion/nexus/crypto"
	"github.com/apex-fusion/nexus/helper/tests"
	"github.com/apex-fusion/nexus/network"
	"github.com/apex-fusion/nexus/secrets"
	"github.com/apex-fusion/nexus/secrets/local"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/hashicorp/go-hclog"
	"github.com/libp2p/go-libp2p/core/peer"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"time"
)

type InitIBFTResult struct {
	Address string
	NodeID  string
}

func (t *TestServer) initNexus() (*InitIBFTResult, error) {
	localSecretsManager, factoryErr := local.SecretsManagerFactory(
		nil,
		&secrets.SecretsManagerParams{
			Logger: hclog.NewNullLogger(),
			Extra: map[string]interface{}{
				secrets.Path: t.Config.DataDir(),
			},
		})
	if factoryErr != nil {
		return nil, factoryErr
	}

	// Generate the IBFT validator private key
	validatorKey, validatorKeyEncoded, keyErr := crypto.GenerateAndEncodeECDSAPrivateKey()
	if keyErr != nil {
		return nil, keyErr
	}

	// Write the validator private key to the secrets manager storage
	if setErr := localSecretsManager.SetSecret(secrets.ValidatorKey, validatorKeyEncoded); setErr != nil {
		return nil, setErr
	}

	// Generate the libp2p private key
	libp2pKey, libp2pKeyEncoded, keyErr := network.GenerateAndEncodeLibp2pKey()
	if keyErr != nil {
		return nil, keyErr
	}

	// Write the networking private key to the secrets manager storage
	if setErr := localSecretsManager.SetSecret(secrets.NetworkKey, libp2pKeyEncoded); setErr != nil {
		return nil, setErr
	}

	// Get the node ID from the private key
	nodeID, err := peer.IDFromPrivateKey(libp2pKey)
	if err != nil {
		return nil, err
	}

	// Template the Nexus secrets config
	secretsConfigPath := filepath.Join(t.Config.RootDir, t.Config.IBFTDir, "secrets.json")
	err = templateFile("nexus-secrets.json", secretsConfigPath, struct{ DataDir string }{DataDir: t.Config.IBFTDir})
	if err != nil {
		return nil, err
	}

	return &InitIBFTResult{
		Address: crypto.PubKeyToAddress(&validatorKey.PublicKey).String(),
		NodeID:  nodeID.String(),
	}, nil
}

func (t *TestServer) templateNexusConfig() error {
	outputFile := filepath.Join(t.Config.RootDir, t.Config.IBFTDir, "config.yaml")

	return templateFile("nexus-config.yaml", outputFile, struct {
		ExecutionGenesisHash string
		EnginePort           int
		GRPCPort             int
		LibP2PPort           int
		PathToJWT            string
		PathToGenesis        string
		DataDir              string
	}{
		ExecutionGenesisHash: t.Config.ExecutionGenesisBlockHash,
		EnginePort:           t.Config.EnginePort,
		LibP2PPort:           t.Config.LibP2PPort,
		PathToJWT:            path.Join(t.Config.RootDir, t.Config.GethDataDir, "jwt.hex"),
		PathToGenesis:        path.Join(t.Config.RootDir, "genesis.json"),
		DataDir:              path.Join(t.Config.RootDir, t.Config.IBFTDir),
		GRPCPort:             t.Config.GRPCPort,
	})
}

func (t *TestServer) startNexus(ctx context.Context) error {
	serverCmd := server.GetCommand()
	args := []string{
		serverCmd.Use,
		"--config", filepath.Join(t.Config.IBFTDir, "config.yaml"),
	}

	if t.Config.BlockTime != 0 {
		args = append(args, "--block-time", strconv.FormatUint(t.Config.BlockTime, 10))
	}

	t.ReleaseReservedPorts()

	// Start the Nexus server
	t.nexusCmd = exec.Command(t.resolveNexusBinary(), args...)
	t.nexusCmd.Dir = t.Config.RootDir

	stdout := t.GetStdout("nexus")
	t.nexusCmd.Stdout = stdout
	t.nexusCmd.Stderr = stdout

	if err := t.nexusCmd.Start(); err != nil {
		return err
	}

	_, err := tests.RetryUntilTimeout(ctx, func() (interface{}, bool) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if _, err := t.Operator().GetStatus(ctx, &empty.Empty{}); err != nil {
			return nil, true
		}

		return nil, false
	})
	if err != nil {
		return err
	}

	// query the chain id
	chainID, err := t.JSONRPC().Eth().ChainID()
	if err != nil {
		return err
	}

	t.chainID = chainID

	return nil
}

func (t *TestServer) generateNexusGenesis() error {
	genesisCmd := genesis.GetCommand()
	args := []string{
		genesisCmd.Use,
		"--consensus", "ibft",
	}

	args = append(args, "--ibft-validators-prefix-path", t.Config.IBFTDirPrefix)

	if t.Config.EpochSize != 0 {
		args = append(args, "--epoch-size", strconv.FormatUint(t.Config.EpochSize, 10))
	}

	for _, bootnode := range t.Config.Bootnodes {
		args = append(args, "--bootnode", bootnode)
	}

	cmd := exec.Command(t.resolveNexusBinary(), args...)
	cmd.Dir = t.Config.RootDir

	stdout := t.GetStdout("nexus-genesis")
	cmd.Stdout = stdout
	cmd.Stderr = stdout

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to generate genesis.json: %w", err)
	}

	// Update the genesis.json file to replace the Params.ForksInTime object
	genesisFilePath := filepath.Join(t.Config.RootDir, "genesis.json")

	// Read the contents of the genesis.json file
	genesisFileContents, err := os.ReadFile(genesisFilePath)
	if err != nil {
		return fmt.Errorf("failed to read genesis.json: %w", err)
	}

	// Parse the JSON into a map for modification
	var genesisData map[string]interface{}
	if err := json.Unmarshal(genesisFileContents, &genesisData); err != nil {
		return fmt.Errorf("failed to parse genesis.json: %w", err)
	}

	// Access Params field and update ForksInTime
	if params, ok := genesisData["params"].(map[string]interface{}); ok {
		params["forks"] = t.Config.Forks
	} else {
		return fmt.Errorf("Params object not found in genesis.json")
	}

	// Marshal the updated data back to JSON
	updatedGenesisFileContents, err := json.MarshalIndent(genesisData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode updated genesis.json: %w", err)
	}

	// Write the updated JSON back to the genesis.json file
	if err := os.WriteFile(genesisFilePath, updatedGenesisFileContents, 0644); err != nil {
		return fmt.Errorf("failed to write updated genesis.json: %w", err)
	}

	return nil
}

func (t *TestServer) resolveNexusBinary() string {
	bin := os.Getenv("NEXUS_BINARY")
	if bin != "" {
		return bin
	}

	binName := "nexus"
	if t.Config.CustomNexusBinary != "" {
		binName = t.Config.CustomNexusBinary
	}

	return fmt.Sprintf("%s/%s", t.Config.RootDir, binName)
}
