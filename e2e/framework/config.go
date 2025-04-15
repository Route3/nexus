package framework

import (
	"crypto/ecdsa"
	"math/big"
	"path/filepath"

	"github.com/apex-fusion/nexus/consensus/ibft"
	"github.com/apex-fusion/nexus/crypto"
	"github.com/apex-fusion/nexus/types"
)

type ConsensusType int

type SrvAccount struct {
	Addr    types.Address
	Balance *big.Int
}

// TestServerConfig for the test server
type TestServerConfig struct {
	ReservedPorts             []ReservedPort
	JWTHex                    string
	GethGenesis               string
	JSONRPCPort               int                  // The geth's JSON RPC endpoint port
	GRPCPort                  int                  // The nexus's GRPC endpoint port
	LibP2PPort                int                  // The nexus's LibP2P endpoint port
	EnginePort                int                  // The geth's Engine API port
	DevP2PPort                int                  // The geth's DevP2P port
	RootDir                   string               // The root directory for test environment
	IBFTDirPrefix             string               // The prefix of data directory for IBFT
	IBFTDir                   string               // The name of data directory for IBFT
	GethDataDir               string               // The name of the data directory for Geth
	Bootnodes                 []string             // Bootnode Addresses
	EpochSize                 uint64               // The epoch size in blocks for the IBFT layer
	ShowsLog                  bool                 // Flag specifying if logs are shown
	Name                      string               // Name of the server
	LogsDir                   string               // Directory where logs are saved
	Signer                    *crypto.EIP155Signer // Signer used for transactions
	BlockTime                 uint64               // Minimum block generation time (in s)
	ExecutionGenesisBlockHash string
}

// DataDir returns path of data directory server uses
func (t *TestServerConfig) DataDir() string {
	return filepath.Join(t.RootDir, t.IBFTDir)
}

func (t *TestServerConfig) SetSigner(signer *crypto.EIP155Signer) {
	t.Signer = signer
}

func (t *TestServerConfig) SetBlockTime(blockTime uint64) {
	t.BlockTime = blockTime
}

// PrivateKey returns a private key in data directory
func (t *TestServerConfig) PrivateKey() (*ecdsa.PrivateKey, error) {
	return crypto.GenerateOrReadPrivateKey(filepath.Join(t.DataDir(), "consensus", ibft.IbftKeyName))
}

// CALLBACKS //

// SetIBFTDir callback sets the name of data directory for IBFT
func (t *TestServerConfig) SetIBFTDir(ibftDir string) {
	t.IBFTDir = ibftDir
}

// SetBootnodes sets bootnodes
func (t *TestServerConfig) SetBootnodes(bootnodes []string) {
	t.Bootnodes = bootnodes
}

// SetShowsLog sets flag for logging
func (t *TestServerConfig) SetShowsLog(f bool) {
	t.ShowsLog = f
}

// SetEpochSize sets the epoch size for the consensus layer.
// It controls the rate at which the validator set is updated
func (t *TestServerConfig) SetEpochSize(epochSize uint64) {
	t.EpochSize = epochSize
}

// SetLogsDir sets the directory where logs are saved
func (t *TestServerConfig) SetLogsDir(dir string) {
	t.LogsDir = dir
}

// SetName sets the name of the server
func (t *TestServerConfig) SetName(name string) {
	t.Name = name
}

// SetGethDataDir sets the name of the data directory for Geth e.g. 1
func (t *TestServerConfig) SetGethDataDir(dataDir string) {
	t.GethDataDir = dataDir
}
