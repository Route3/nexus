package server

import (
	"net"

	"github.com/hashicorp/go-hclog"

	"github.com/apex-fusion/nexus/chain"
	"github.com/apex-fusion/nexus/engine"
	"github.com/apex-fusion/nexus/network"
	"github.com/apex-fusion/nexus/secrets"
)

const (
	DefaultGRPCPort    int = 9632
	DefaultJSONRPCPort int = 8545
)

// Config is used to parametrize the minimal client
type Config struct {
	Chain *chain.Chain

	JSONRPC    *JSONRPC
	GRPCAddr   *net.TCPAddr
	LibP2PAddr *net.TCPAddr

	PriceLimit         uint64
	MaxAccountEnqueued uint64
	MaxSlots           uint64
	BlockTime          uint64

	Telemetry *Telemetry
	Network   *network.Config

	DataDir     string
	RestoreFile *string

	Seal bool

	SecretsManager *secrets.SecretsManagerConfig

	LogLevel hclog.Level

	JSONLogFormat bool

	LogFilePath string

	EngineConfig engine.EngineConfig

	ExecutionGenesisHash string
	SuggestedFeeRecipient string
}

// Telemetry holds the config details for metric services
type Telemetry struct {
	PrometheusAddr *net.TCPAddr
}

// JSONRPC holds the config details for the JSON-RPC server
type JSONRPC struct {
	JSONRPCAddr              *net.TCPAddr
	AccessControlAllowOrigin []string
	BatchLengthLimit         uint64
	BlockRangeLimit          uint64
}
