package main

import (
	"context"
	"dagger/gnogenesis/internal/dagger"
	"fmt"
	"strings"
)

// use entrypoint
var execOpts = dagger.ContainerWithExecOpts{
	UseEntrypoint: true,
}

type Gnogenesis struct{}

type GitLocator dagger.GitclonerLocator

// Gathers the built binary according to target
func (m *Gnogenesis) getBinary(target dagger.GitclonerTargetBinary, sourceOpts *dagger.GitclonerBuildImageFromSourceOpts) *dagger.Container {
	if sourceOpts == nil {
		return dag.Container().
			From(m.getMasterImage(target))
	}

	return dag.Gitcloner().BuildImageFromSource(target, *sourceOpts)
}

// Gathers the name of the latest image according to the target
func (r *Gnogenesis) getMasterImage(target dagger.GitclonerTargetBinary) string {
	switch target {
	case dagger.GitclonerTargetBinaryGnocontribsBin,
		dagger.GitclonerTargetBinaryGnokeyBin,
		dagger.GitclonerTargetBinaryGnolandBin:
		return fmt.Sprintf("ghcr.io/gnolang/gno/%s:master", target)
	default:
		fmt.Println(target)
	}
	return ""
}

// Generates a genesis file using `gnogenesis` binary
func (m *Gnogenesis) generate(sourceOpts *dagger.GitclonerBuildImageFromSourceOpts) *dagger.File {
	return m.getBinary(dagger.GitclonerTargetBinaryGnocontribsBin, sourceOpts).
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

// Verify a genesis using Gnogeneis master image
func (m *Gnogenesis) Verify(ctx context.Context) (int, error) {
	secretsCtr := m.generateBasicNode(nil)
	genesisWithValnode := m.AddValidatorNode(ctx,
		"val0",
		secretsCtr.File("/gnoroot/genesis.json"),
		secretsCtr.Directory("/gnoroot/gnoland-data/secrets/"))

	return m.getBinary(dagger.GitclonerTargetBinaryGnocontribsBin, nil).
		WithWorkdir("/gnoroot").
		WithFile("/gnoroot/genesis.json", genesisWithValnode).
		WithExec([]string{"gnogenesis", "verify"}).
		ExitCode(ctx)
}

// Runs a Gnoland Binary using a generated genesis
func (m *Gnogenesis) generateBasicNode(
	sourceOpts *dagger.GitclonerBuildImageFromSourceOpts) *dagger.Container {
	return m.getBinary(dagger.GitclonerTargetBinaryGnolandBin, sourceOpts).
		WithExec([]string{"config", "init"}, execOpts).
		WithExec([]string{"secrets", "init"}, execOpts).
		WithFile("/gnoroot/genesis.json", m.generate(sourceOpts))
}

// Runs a Gnoland Binary using a generated genesis
func (m *Gnogenesis) runGnolandWithGenesis(
	ctx context.Context,
	sourceOpts *dagger.GitclonerBuildImageFromSourceOpts) *dagger.Container {
	secretsCtr := m.generateBasicNode(sourceOpts)

	moniker, _ := secretsCtr.WithExec([]string{"config", "get", "moniker"}, execOpts).Stdout(ctx)

	// add THIS node as validator
	genesisWithValnode := m.AddValidatorNode(ctx,
		moniker,
		secretsCtr.File("/gnoroot/genesis.json"),
		secretsCtr.Directory("/gnoroot/gnoland-data/secrets/"))

	return m.getBinary(dagger.GitclonerTargetBinaryGnolandBin, sourceOpts).
		WithFile("/gnoroot/genesis.json", genesisWithValnode).
		WithDirectory("/gnoroot/gnoland-data", secretsCtr.Directory("/gnoroot/gnoland-data")).
		WithExec([]string{"start", "-genesis=/gnoroot/genesis.json", "-log-level=info", "-skip-genesis-sig-verification"}, execOpts)
}

// Runs a Gnoland Binary using master image
func (m *Gnogenesis) RunGnolandWithGenesis(ctx context.Context) *dagger.Container {
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
) *dagger.Container {
	srcDir := &dagger.GitclonerBuildImageFromSourceOpts{
		Locator: dagger.GitclonerLocator(locator),
		Ref:     ref,
		Fork:    fork,
	}

	return m.runGnolandWithGenesis(ctx, srcDir)
}

// Gets node id from secret folder
func (m *Gnogenesis) GetNodeId(ctx context.Context, secretsFolder *dagger.Directory) (string, error) {
	nodeId, err := m.getBinary(dagger.GitclonerTargetBinaryGnolandBin, nil).
		WithDirectory("/gnoroot/gnoland-data/secrets", secretsFolder).
		WithExec(strings.Split("secrets get node_id.id -raw", " "), execOpts).Stdout(ctx)

	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(nodeId, "\n", ""), nil
}

// Adds a validator node to the current genesis
func (m *Gnogenesis) AddValidatorNode(
	ctx context.Context,
	nodeName string,
	genesisFile *dagger.File,
	secretsFolder *dagger.Directory) *dagger.File {
	// get node address
	nodeAddress, _ := m.getBinary(dagger.GitclonerTargetBinaryGnolandBin, nil).
		WithDirectory("/gnoroot/gnoland-data/secrets", secretsFolder).
		WithExec(strings.Split("secrets get validator_key.address -raw", " "), execOpts).Stdout(ctx)

	// get node pub key
	nodePubKey, _ := m.getBinary(dagger.GitclonerTargetBinaryGnolandBin, nil).
		WithDirectory("/gnoroot/gnoland-data/secrets", secretsFolder).
		WithExec(strings.Split("secrets get validator_key.pub_key -raw", " "), execOpts).Stdout(ctx)

	return m.getBinary(dagger.GitclonerTargetBinaryGnocontribsBin, nil).
		WithWorkdir("/gnoroot").
		WithFile("/gnoroot/genesis.json", genesisFile).
		WithExec([]string{
			"gnogenesis",
			"validator",
			"add",
			"-name",
			nodeName,
			"-address",
			strings.ReplaceAll(nodeAddress, "\n", ""),
			"-pub-key",
			strings.ReplaceAll(nodePubKey, "\n", ""),
		}).
		File("/gnoroot/genesis.json")
}
