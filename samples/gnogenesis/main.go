package main

import (
	"context"
	"dagger/gno-dagger/internal/dagger"
	"strings"
)

type Gnogenesis struct{}

// Generates a genesis file
func (m *Gnogenesis) Generate(ctx context.Context) *dagger.File {
	return dag.Container().
		From("ghcr.io/gnolang/gno/gnocontribs:0.0.1-6bf3889b-master").
		WithDirectory("/src", dag.Gitcloner().Clone(dagger.GitclonerCloneOpts{
			Ref:     "6bf3889b69bd46139abe8a308beec3deda679d92",
			Locator: dagger.GitclonerLocatorCommit,
		})).
		WithExec([]string{"gnogenesis", "generate"}).
		WithExec(strings.Split("gnogenesis txs add packages /src/examples", " ")).
		WithExec(strings.Split("gnogenesis balances add -balance-sheet /src/gno.land/genesis/genesis_balances.txt", " ")).
		File("/gnoroot/genesis.json")
}

// Runs a Gnoland Binary using a generated genesis
func (m *Gnogenesis) RunGnolandWithGenesis(ctx context.Context) (int, error) {
	// use entrypoint
	execOpts := dagger.ContainerWithExecOpts{
		UseEntrypoint: true,
	}

	return dag.Container().
		From("ghcr.io/gnolang/gno/gnoland:0.0.1-6bf3889b-master").
		WithExec([]string{"config", "init"}, execOpts).
		WithExec([]string{"secrets", "init"}, execOpts).
		WithFile("/gnoroot/genesis.json", m.Generate(ctx)).
		Terminal().
		WithExec([]string{"start", "--genesis", "/gnoroot/genesis.json", "--log-level", "info"}, execOpts).
		ExitCode(ctx)
}
