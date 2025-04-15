package e2e

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
	"time"

	"github.com/apex-fusion/nexus/e2e/framework"
)

func TestClusterBlockSync(t *testing.T) {
	const (
		numNonValidators = 2
		numberOfServers  = IBFTMinNodes + numNonValidators
		desiredHeight    = 10
	)

	t.Helper()

	// Start IBFT cluster (4 Validator + 2 Non-Validator)
	ibftManager, err := framework.NewServerManager(
		t,
		numberOfServers,
		func(i int, config *framework.TestServerConfig) {
			if i >= IBFTMinNodes {
				// Other nodes should not be in the validator set
				dirPrefix := "nexus-non-validator-"
				config.SetIBFTDir(fmt.Sprintf("%s%d", dirPrefix, i))
			}
		})

	require.NoError(t, err)

	startContext, startCancelFn := context.WithTimeout(context.Background(), time.Minute)
	defer startCancelFn()
	ibftManager.StartServers(startContext)

	servers := make([]*framework.TestServer, 0)
	for i := 0; i < numberOfServers; i++ {
		servers = append(servers, ibftManager.GetServer(i))
	}
	// All nodes should have mined the same block eventually
	waitErrors := framework.WaitForServersToSeal(servers, desiredHeight)

	if len(waitErrors) != 0 {
		t.Fatalf("Unable to wait for all nodes to seal blocks, %v", waitErrors)
	}
}
