package tests

import (
	"strconv"
	"testing"
)

func TestE2EMultiBroadcast(t *testing.T) {

	//test setup
	cfg := LoadMultiTestConfig()

	clts, _ := basicMultiSetup(t)

	//test suite
	for idx, clt := range clts {

		testBroadcastTx(strconv.Itoa(idx), t, clt, cfg.masterAccountPrivateKey, cfg.rpcUrls[0])
		defaultDelay()
		defaultDelay()
	}

	//test teardown
	t.Cleanup(func() { cleanupDockerEnv(t) })
}
