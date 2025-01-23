package tests

import (
	"strconv"
	"testing"
)

func TestE2EMultiLiveness(t *testing.T) {

	//test setup
	clts, masterAcc := basicMultiSetup(t)

	//test suite
	for idx, clt := range clts {

		vId := strconv.Itoa(idx)

		testFetchAndCheckMetaFields(vId, t, clt)

		testBalanceGreaterThanZero(vId, t, clt, masterAcc)

		testBlockAreBeingProduced(vId, t, clt)
	}

	//test teardown
	t.Cleanup(func() { cleanupDockerEnv(t) })
}
