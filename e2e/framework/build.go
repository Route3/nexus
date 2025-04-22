package framework

import (
	"os/exec"
)

func Build() error {
	// Compile without optimizations and inlining
	args := []string{
		"build",
		"-gcflags=all=-N -l",
		"-o", "e2e/framework/artifacts/nexus",
	}

	buildCommand := exec.Command("go", args...)
	buildCommand.Dir = "../"

	return buildCommand.Run()
}
