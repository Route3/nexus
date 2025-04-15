package framework

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/apex-fusion/nexus/helper/tests"
)

var (
	DefaultTimeout = 5 * time.Minute
)

type AtomicErrors struct {
	sync.RWMutex
	errors []error
}

func (a *AtomicErrors) Append(err error) {
	a.Lock()
	defer a.Unlock()

	a.errors = append(a.errors, err)
}

func (a *AtomicErrors) Errors() []error {
	a.RLock()
	defer a.RUnlock()

	return a.errors
}

// WaitUntilBlockMined waits until server mined block with bigger height than given height
// otherwise returns timeout
func WaitUntilBlockMined(ctx context.Context, srv *TestServer, desiredHeight uint64) (uint64, error) {
	clt := srv.JSONRPC().Eth()
	res, err := tests.RetryUntilTimeout(ctx, func() (interface{}, bool) {
		height, err := clt.BlockNumber()
		if err == nil && height >= desiredHeight {
			return height, false
		}

		return nil, true
	})

	if err != nil {
		return 0, err
	}

	blockNum, ok := res.(uint64)
	if !ok {
		return 0, errors.New("invalid type assert")
	}

	return blockNum, nil
}

// tempDir returns directory path in tmp with random directory name
func tempDir() (string, error) {
	return os.MkdirTemp("/tmp", "nexus-e2e-")
}

func ToLocalIPv4LibP2pAddr(port int, nodeID string) string {
	return fmt.Sprintf("/ip4/127.0.0.1/tcp/%d/p2p/%s", port, nodeID)
}

// ReservedPort keeps available port until use
type ReservedPort struct {
	port     int
	listener net.Listener
	isClosed bool
}

func (p *ReservedPort) Port() int {
	return p.port
}

func (p *ReservedPort) IsClosed() bool {
	return p.isClosed
}

func (p *ReservedPort) Close() error {
	if p.isClosed {
		return nil
	}

	err := p.listener.Close()
	p.isClosed = true

	return err
}

func FindAvailablePort(from, to int) *ReservedPort {
	for port := from; port < to; port++ {
		addr := fmt.Sprintf("localhost:%d", port)
		if l, err := net.Listen("tcp", addr); err == nil {
			return &ReservedPort{port: port, listener: l}
		}
	}

	return nil
}

func FindAvailablePorts(n, from, to int) ([]ReservedPort, error) {
	ports := make([]ReservedPort, 0, n)
	nextFrom := from

	for i := 0; i < n; i++ {
		newPort := FindAvailablePort(nextFrom, to)
		if newPort == nil {
			// Close current reserved ports
			for _, p := range ports {
				err := p.Close()
				if err != nil {
					return nil, err
				}
			}

			return nil, errors.New("couldn't reserve required number of ports")
		}

		ports = append(ports, *newPort)
		nextFrom = newPort.Port() + 1
	}

	return ports, nil
}

func WaitForServersToSeal(servers []*TestServer, desiredHeight uint64) []error {
	waitErrors := make([]error, 0)

	var waitErrorsLock sync.Mutex

	appendWaitErr := func(waitErr error) {
		waitErrorsLock.Lock()
		defer waitErrorsLock.Unlock()

		waitErrors = append(waitErrors, waitErr)
	}

	var wg sync.WaitGroup
	for i := 0; i < len(servers); i++ {
		wg.Add(1)

		go func(i int) {
			waitCtx, waitCancelFn := context.WithTimeout(context.Background(), time.Minute)
			defer func() {
				waitCancelFn()
				wg.Done()
			}()

			_, waitErr := WaitUntilBlockMined(waitCtx, servers[i], desiredHeight)
			if waitErr != nil {
				appendWaitErr(fmt.Errorf("unable to wait for block, %w", waitErr))
			}
		}(i)
	}
	wg.Wait()

	return waitErrors
}
