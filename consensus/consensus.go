package consensus

import (
	"context"
	"log"

	"github.com/apex-fusion/nexus/blockchain"
	"github.com/apex-fusion/nexus/chain"
	"github.com/apex-fusion/nexus/helper/progress"
	"github.com/apex-fusion/nexus/network"
	"github.com/apex-fusion/nexus/secrets"
	"github.com/apex-fusion/nexus/types"
	"github.com/hashicorp/go-hclog"
	"google.golang.org/grpc"
)

// Consensus is the public interface for consensus mechanism
// Each consensus mechanism must implement this interface in order to be valid
type Consensus interface {
	// VerifyHeader verifies the header is correct
	VerifyHeader(header *types.Header) error

	// ProcessHeaders updates the snapshot based on the verified headers
	ProcessHeaders(headers []*types.Header) error

	// GetBlockCreator retrieves the block creator (or signer) given the block header
	GetBlockCreator(header *types.Header) (types.Address, error)

	// PreCommitState a hook to be called before finalizing state transition on inserting block
	PreCommitState(header *types.Header) error

	// GetSyncProgression retrieves the current sync progression, if any
	GetSyncProgression() *progress.Progression

	// Initialize initializes the consensus (e.g. setup data)
	Initialize() error

	// Start starts the consensus and servers
	Start() error

	// Close closes the connection
	Close() error
}

// Config is the configuration for the consensus
type Config struct {
	// Logger to be used by the consensus
	Logger *log.Logger

	// Params are the params of the chain and the consensus
	Params *chain.Params

	// Config defines specific configuration parameters for the consensus
	Config map[string]interface{}

	// Path is the directory path for the consensus protocol tos tore information
	Path string
}

type Params struct {
	Context        context.Context
	Config         *Config
	Network        *network.Server
	Blockchain     *blockchain.Blockchain
	Grpc           *grpc.Server
	Logger         hclog.Logger
	SecretsManager secrets.SecretsManager
	BlockTime      uint64
}

// Factory is the factory function to create a discovery consensus
type Factory func(*Params) (Consensus, error)
