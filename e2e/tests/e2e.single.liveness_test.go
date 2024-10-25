package tests

import (
	"fmt"
	"testing"
)

func TestE2ESingleLiveness(t *testing.T) {

	fmt.Println("Start:TestE2ESingleLiveness")

	//test setup
	clt, masterAcc := basicSingleSetup(t)

	//test suite
	testFetchAndCheckMetaFields("0", t, clt)

	testBalanceGreaterThanZero("0", t, clt, masterAcc)

	testBlockAreBeingProduced("0", t, clt)
	
	//test teardown
	t.Cleanup(func() { cleanupDockerEnv(t) })
}
