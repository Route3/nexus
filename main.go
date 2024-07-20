package main

import (
	_ "embed"

	"github.com/apex-fusion/nexus/command/root"
	"github.com/apex-fusion/nexus/licenses"
)

var (
	//go:embed LICENSE
	license string
)

func main() {
	licenses.SetLicense(license)

	root.NewRootCommand().Execute()
}
