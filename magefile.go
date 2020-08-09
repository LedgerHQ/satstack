// +build mage

package main

import (
	"os"

	"github.com/magefile/mage/sh"
)

const (
	entryPoint = "cmd/lss.go"
	ldFlags    = "-X '$PACKAGE/version.GitCommit=$COMMIT_HASH' " +
		"-X '$PACKAGE/version.Build=$BUILD'"
)

// Allow user to override Go executable on UNIX-like systems.
var goexe = "go" // GOEXE=xxx mage build

func init() {
	if exe := os.Getenv("GOEXE"); exe != "" {
		goexe = exe
	}

	// We want to use Go 1.11 modules even if the source lives inside GOPATH.
	// The default is "auto".
	os.Setenv("GO111MODULE", "on")
}

// Build binary
func Build() error {
	return sh.RunWith(flagEnv(), goexe, "build", "-ldflags", ldFlags,
		entryPoint)
}

func Release() error {
	os.Setenv("GIN_MODE", "release")
	return Build()
}

// Run basic golangci-lint check.
func Lint() error {
	linterArgs := []string{
		"run",
		"--disable-all",
		"--enable=govet",
		"--enable=gofmt",
		"--enable=gosec",
	}

	if err := sh.Run("golangci-lint", linterArgs...); err != nil {
		return err
	}

	return nil
}

func flagEnv() map[string]string {
	hash, _ := sh.Output("git", "rev-parse", "--short", "HEAD")

	build := "development"
	if mode := os.Getenv("GIN_MODE"); mode == "release" {
		build = "release"
	}

	return map[string]string{
		"PACKAGE":     "ledger-sats-stack",
		"COMMIT_HASH": hash,
		"BUILD":       build,
	}
}
