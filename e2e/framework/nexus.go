package framework

import (
	"context"
	"fmt"
	"github.com/apex-fusion/nexus/command/genesis"
	"github.com/apex-fusion/nexus/command/server"
	"github.com/apex-fusion/nexus/crypto"
	"github.com/apex-fusion/nexus/helper/tests"
	"github.com/apex-fusion/nexus/network"
	"github.com/apex-fusion/nexus/secrets"
	"github.com/apex-fusion/nexus/secrets/local"
	"github.com/apex-fusion/nexus/validators"
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

	if t.Config.ValidatorType == validators.BLSValidatorType {
		// Generate the BLS Key
		_, bksKeyEncoded, keyErr := crypto.GenerateAndEncodeBLSSecretKey()
		if keyErr != nil {
			return nil, keyErr
		}

		// Write the networking private key to the secrets manager storage
		if setErr := localSecretsManager.SetSecret(secrets.ValidatorBLSKey, bksKeyEncoded); setErr != nil {
			return nil, setErr
		}
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
		PathToJWT            string
		PathToGenesis        string
		DataDir              string
	}{
		ExecutionGenesisHash: t.Config.ExecutionGenesisBlockHash,
		EnginePort:           t.Config.EnginePort,
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
		"--ibft-validator-type", string(t.Config.ValidatorType),
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
	return err
}

func (t *TestServer) resolveNexusBinary() string {
	bin := os.Getenv("NEXUS_BINARY")
	if bin != "" {
		return bin
	}
	// fallback
	return fmt.Sprintf("%s/%s", t.Config.RootDir, "nexus")
}
