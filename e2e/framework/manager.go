package framework

import (
	"context"
	"fmt"
	"os"
	"path"
	"testing"
	"time"
)

// generated using https://vanity-eth.tk/
// solely for testing purposes!
const PREMINE_ADDRESS = "0x735773c4C940b849D457aDCf0e519D75d384Af27"
const PREMINE_PRIVATE_KEY = "3e07b681a2d6b12b76e41b62fabccd8966264066c2d79285fd13e0de646e44f0"
const PREMINE_BALANCE = "0x9B18AB5DF7180B6B8000000"

type ServerManager struct {
	t                  *testing.T
	servers            []*TestServer
	PremineAllocations map[string]string // In hex format the premine alloc for EL
}

type IBFTServerConfigCallback func(index int, config *TestServerConfig)

var startTime int64

func init() {
	startTime = time.Now().UnixMilli()
}

func NewServerManager(
	t *testing.T,
	numNodes int,
	callback IBFTServerConfigCallback,
) (*ServerManager, error) {
	t.Helper()

	dataDir, err := tempDir()
	if err != nil {
		return nil, err
	}

	err = setupBinaries(dataDir)
	if err != nil {
		return nil, err
	}

	premineAllocations := make(map[string]string)
	premineAllocations[PREMINE_ADDRESS] = PREMINE_BALANCE
	err = templateGethGenesis(fmt.Sprintf("%s/geth-genesis.json", dataDir), premineAllocations)
	if err != nil {
		return nil, err
	}

	servers := make([]*TestServer, 0, numNodes)

	t.Cleanup(func() {
		for _, s := range servers {
			s.Stop()
		}
		if err := os.RemoveAll(dataDir); err != nil {
			t.Log(err)
		}
	})

	bootnodes := make([]string, 0, numNodes)
	genesisValidators := make([]string, 0, numNodes)

	logsDir, err := initLogsDir(t)
	if err != nil {
		return nil, err
	}

	for i := 0; i < numNodes; i++ {
		srv := NewTestServer(t, dataDir, func(config *TestServerConfig) {
			config.SetIBFTDir(fmt.Sprintf("%s%d", config.IBFTDirPrefix, i))
			config.SetGethDataDir(fmt.Sprintf("%s%d", "e2e-geth-", i))
			config.SetLogsDir(logsDir)
			config.SetName(fmt.Sprintf("node-%d", i))
			callback(i, config)
		})

		res, err := srv.initNexus()
		if err != nil {
			return nil, err
		}

		err = srv.initGeth()
		if err != nil {
			return nil, err
		}

		libp2pAddr := ToLocalIPv4LibP2pAddr(srv.Config.LibP2PPort, res.NodeID)

		servers = append(servers, srv)
		bootnodes = append(bootnodes, libp2pAddr)
		genesisValidators = append(genesisValidators, res.Address)
	}

	srv := servers[0]
	srv.Config.SetBootnodes(bootnodes)

	if err := srv.generateNexusGenesis(); err != nil {
		return nil, err
	}

	return &ServerManager{t, servers, premineAllocations}, nil
}

func (m *ServerManager) StartServers(ctx context.Context) {
	for idx, srv := range m.servers {
		if err := srv.Start(ctx); err != nil {
			m.t.Fatal(fmt.Errorf("server %d failed to start: %+v", idx, err))
		}
	}

	for idx, srv := range m.servers {
		if err := srv.WaitForReady(ctx); err != nil {
			m.t.Logf("server %d couldn't advance block: %+v", idx, err)
			m.t.Fatal(err)
		}
	}
}

func (m *ServerManager) StopServers() {
	for _, srv := range m.servers {
		srv.Stop()
	}
}

func (m *ServerManager) GetServer(i int) *TestServer {
	if i >= len(m.servers) {
		return nil
	}

	return m.servers[i]
}

func initLogsDir(t *testing.T) (string, error) {
	t.Helper()
	logsDir := path.Join("..", "e2e-logs", fmt.Sprintf("e2e-logs-%d", startTime), t.Name())

	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return "", err
	}

	return logsDir, nil
}
