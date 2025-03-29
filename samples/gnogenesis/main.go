package main

import (
	"context"
	"dagger/gno-dagger/internal/dagger"
	"fmt"
	"strings"
)

type Gnogenesis struct{}

type GitLocator dagger.GitclonerLocator

// Gathers the built binary  according to target
func (m *Gnogenesis) getBinary(target dagger.GitclonerTargetBinary, sourceOpts *dagger.GitclonerBuildImageFromSourceOpts) *dagger.Container {
	if sourceOpts == nil {
		return dag.Container().
			From(m.getMasterImage(target))
	}

	return dag.Gitcloner().BuildImageFromSource(target,
		dagger.GitclonerBuildImageFromSourceOpts{
			Locator: dagger.GitclonerLocatorCommit,
			Ref:     "632a38a7ba3b9b88cd85cd8b345f215d9015fdca",
			Fork:    "aeddi",
		})
}

// Gathers the name of the latest image according to the target
func (r *Gnogenesis) getMasterImage(target dagger.GitclonerTargetBinary) string {
	switch target {
	case dagger.GitclonerTargetBinaryGnocontribs,
		dagger.GitclonerTargetBinaryGnokey,
		dagger.GitclonerTargetBinaryGnoland:
		fmt.Println("ok")
		return fmt.Sprintf("ghcr.io/gnolang/gno/%s:master", target)
	default:
		fmt.Println(target)
	}
	return ""
}

// Generates a genesis file using `gnogenesis` binary
func (m *Gnogenesis) generate(sourceOpts *dagger.GitclonerBuildImageFromSourceOpts) *dagger.File {
	return m.getBinary(dagger.GitclonerTargetBinaryGnocontribs, sourceOpts).
		WithWorkdir("/gnoroot").
		WithExec([]string{"gnogenesis", "generate"}).
		WithExec(strings.Split("gnogenesis balances add -balance-sheet /gnoroot/gno.land/genesis/genesis_balances.txt", " ")).
		WithExec(strings.Split("gnogenesis txs add packages /gnoroot/examples", " ")).
		File("/gnoroot/genesis.json")
}

// Generates a genesis file starting from a binary generated from Gnoland source code
func (m *Gnogenesis) GenerateUsingCodebase(
	// +optional
	locator GitLocator,
	// +optional
	ref string,
	// +optional
	fork string,
) *dagger.File {
	srcDir := &dagger.GitclonerBuildImageFromSourceOpts{
		Locator: dagger.GitclonerLocator(locator),
		Ref:     ref,
		Fork:    fork,
	}

	return m.generate(srcDir)
}

// Generates a genesis using Gnogeneis master image
func (m *Gnogenesis) Generate() *dagger.File {
	return m.generate(nil)
}

// Runs a Gnoland Binary using a generated genesis
func (m *Gnogenesis) runGnolandWithGenesis(
	ctx context.Context,
	sourceOpts *dagger.GitclonerBuildImageFromSourceOpts) (int, error) {
	// use entrypoint
	execOpts := dagger.ContainerWithExecOpts{
		UseEntrypoint: true,
	}

	return m.getBinary(dagger.GitclonerTargetBinaryGnoland, sourceOpts).
		WithExec([]string{"config", "init"}, execOpts).
		WithExec([]string{"secrets", "init"}, execOpts).
		WithFile("/gnoroot/genesis.json", m.generate(sourceOpts)).
		// Terminal().
		WithExec([]string{"start", "-genesis=/gnoroot/genesis.json", "-log-level=info"}, execOpts).
		ExitCode(ctx)
}

// Runs a Gnoland Binary using master image
func (m *Gnogenesis) RunGnolandWithGenesis(ctx context.Context) (int, error) {
	return m.runGnolandWithGenesis(ctx, nil)
}

// Runs a Gnoland Binary generated from Gnoland source code
func (m *Gnogenesis) RunGnolandWithGenesisUsingCodebase(
	ctx context.Context,
	// +optional
	locator GitLocator,
	// +optional
	ref string,
	// +optional
	fork string,
) (int, error) {
	srcDir := &dagger.GitclonerBuildImageFromSourceOpts{
		Locator: dagger.GitclonerLocator(locator),
		Ref:     ref,
		Fork:    fork,
	}

	return m.runGnolandWithGenesis(ctx, srcDir)
}
