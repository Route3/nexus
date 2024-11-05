package tests

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/apex-fusion/nexus/chain"
	"github.com/apex-fusion/nexus/types"
)

var (
	stateTests       = "GeneralStateTests"
	legacyStateTests = "LegacyTests/Constantinople/GeneralStateTests"
)

type stateCase struct {
	Info        *info                                   `json:"_info"`
	Env         *env                                    `json:"env"`
	Pre         map[types.Address]*chain.GenesisAccount `json:"pre"`
	Post        map[string]postState                    `json:"post"`
	Transaction *stTransaction                          `json:"transaction"`
}

var ripemd = types.StringToAddress("0000000000000000000000000000000000000003")

func TestState(t *testing.T) {
	t.Parallel()

	long := []string{
		"static_Call50000",
		"static_Return50000",
		"static_Call1MB",
		"stQuadraticComplexityTest",
		"stTimeConsuming",
	}

	skip := []string{
		"RevertPrecompiledTouch",
	}

	// There are two folders in spec tests, one for the current tests for the Istanbul fork
	// and one for the legacy tests for the other forks
	folders, err := listFolders(stateTests, legacyStateTests)
	if err != nil {
		t.Fatal(err)
	}

	for _, folder := range folders {
		folder := folder
		t.Run(folder, func(t *testing.T) {
			t.Parallel()

			files, err := listFiles(folder)
			if err != nil {
				t.Fatal(err)
			}

			for _, file := range files {
				if !strings.HasSuffix(file, ".json") {
					continue
				}

				if contains(long, file) && testing.Short() {
					t.Skipf("Long tests are skipped in short mode")

					continue
				}

				if contains(skip, file) {
					t.Skip()

					continue
				}

				data, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}

				var c map[string]stateCase
				if err := json.Unmarshal(data, &c); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
