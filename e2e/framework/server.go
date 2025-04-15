package framework

import (
	"context"
	"fmt"
	ibftOp "github.com/apex-fusion/nexus/consensus/ibft/proto"
	"github.com/apex-fusion/nexus/crypto"
	"github.com/apex-fusion/nexus/server/proto"
	"github.com/umbracle/ethgo/jsonrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

type TestServerConfigCallback func(*TestServerConfig)

const (
	serverIP    = "127.0.0.1"
	initialPort = 12000
)

type TestServer struct {
	t *testing.T

	Config   *TestServerConfig
	nexusCmd *exec.Cmd
	gethCmd  *exec.Cmd
	chainID  *big.Int
}

func NewTestServer(t *testing.T, rootDir string, callback TestServerConfigCallback) *TestServer {
	t.Helper()

	// Reserve ports
	ports, err := FindAvailablePorts(5, initialPort, initialPort+10000)
	if err != nil {
		t.Fatal(err)
	}

	// Sets the services to start on open ports
	config := &TestServerConfig{
		IBFTDirPrefix: "e2e-nexus-",
		ReservedPorts: ports,
		GRPCPort:      ports[0].Port(),
		LibP2PPort:    ports[1].Port(),
		JSONRPCPort:   ports[2].Port(),
		EnginePort:    ports[3].Port(),
		DevP2PPort:    ports[4].Port(),
		RootDir:       rootDir,
		Signer:        crypto.NewEIP155Signer(100),
	}

	if callback != nil {
		callback(config)
	}

	return &TestServer{
		t:      t,
		Config: config,
	}
}

func (t *TestServer) GrpcAddr() string {
	return fmt.Sprintf("%s:%d", serverIP, t.Config.GRPCPort)
}

func (t *TestServer) LibP2PAddr() string {
	return fmt.Sprintf("%s:%d", serverIP, t.Config.LibP2PPort)
}

func (t *TestServer) JSONRPCAddr() string {
	return fmt.Sprintf("%s:%d", serverIP, t.Config.JSONRPCPort)
}

func (t *TestServer) HTTPJSONRPCURL() string {
	return fmt.Sprintf("http://%s", t.JSONRPCAddr())
}

func (t *TestServer) JSONRPC() *jsonrpc.Client {
	clt, err := jsonrpc.NewClient(t.HTTPJSONRPCURL())
	if err != nil {
		t.t.Fatal(err)
	}

	return clt
}

func (t *TestServer) Operator() proto.SystemClient {
	//goland:noinspection GoDeprecation
	conn, err := grpc.Dial(
		t.GrpcAddr(),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.t.Fatal(err)
	}

	return proto.NewSystemClient(conn)
}

func (t *TestServer) IBFTOperator() ibftOp.IbftOperatorClient {
	//goland:noinspection GoDeprecation
	conn, err := grpc.Dial(
		t.GrpcAddr(),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.t.Fatal(err)
	}

	return ibftOp.NewIbftOperatorClient(conn)
}

func (t *TestServer) ReleaseReservedPorts() {
	for _, p := range t.Config.ReservedPorts {
		if err := p.Close(); err != nil {
			t.t.Error(err)
		}
	}

	t.Config.ReservedPorts = nil
}

func (t *TestServer) Stop() {
	t.ReleaseReservedPorts()

	if t.nexusCmd != nil {
		if err := t.nexusCmd.Process.Kill(); err != nil {
			t.t.Error(err)
		}
	}

	if t.gethCmd != nil {
		if err := t.gethCmd.Process.Kill(); err != nil {
			t.t.Error(err)
		}
	}
}

func (t *TestServer) Start(ctx context.Context, bootnodeEnode string) error {
	err := t.startGeth(ctx, bootnodeEnode)
	if err != nil {
		return err
	}

	// We template the Nexus config now that we have the t.Config.ExecutionGenesisHash
	err = t.templateNexusConfig()
	if err != nil {
		return err
	}

	return t.startNexus(ctx)
}

// GetStdout returns the combined stdout writers of the server
func (t *TestServer) GetStdout(logFilename string) io.Writer {
	var writers []io.Writer

	f, err := os.OpenFile(filepath.Join(t.Config.LogsDir, fmt.Sprintf("%s-%s.log", logFilename, t.Config.Name)), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		t.t.Fatal(err)
	}

	writers = append(writers, f)

	t.t.Cleanup(func() {
		err = f.Close()
		if err != nil {
			t.t.Logf("Failed to close file. Error: %s", err)
		}
	})

	if t.Config.ShowsLog {
		writers = append(writers, os.Stdout)
	}

	if len(writers) == 0 {
		return io.Discard
	}

	return io.MultiWriter(writers...)
}
