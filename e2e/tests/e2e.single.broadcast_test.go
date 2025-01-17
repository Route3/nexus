package tests

import (
	"testing"
)

func TestE2ESingleBroadcast(t *testing.T) {

	//test setup
	cfg := LoadSingleTestConfig()

	clt, _ := basicSingleSetup(t)

	//test suite
	testBroadcastTx("0", t, clt, cfg.masterAccountPrivateKey, cfg.rpcUrls[0])

	//test teardown
	t.Cleanup(func() { cleanupDockerEnv(t) })
}
