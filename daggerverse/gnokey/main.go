package main

import (
	"context"
	"dagger/gnokey/internal/dagger"
	"fmt"
	"strings"
	"time"
)

type Gnokey struct{}

const (
	ChainId   string = "dev"
	KeyId     string = "key0"
	RealmName string = "r/demo/counter"
)

// Gathers Gnokey Container
func (m *Gnokey) BaseGnokey(homeDirKey *dagger.Directory) *dagger.Container {
	return dag.Container().
		From("ghcr.io/gnolang/gno/gnokey:master").
		WithMountedDirectory("/gnohome", homeDirKey)
}

// Generates a Gno key
func (m *Gnokey) GenerateKey(ctx context.Context, homeDirKey *dagger.Directory, passwordString string) *dagger.Container {
	return m.BaseGnokey(homeDirKey).
		WithExec([]string{"gnokey", "add", KeyId, "-home=/gnohome", "-insecure-password-stdin"},
			dagger.ContainerWithExecOpts{
				Stdin: fmt.Sprintf("%[1]s\n%[1]s\n", passwordString),
			})
}

// Performs a Tx using Gnokey on a local chain
func (m *Gnokey) MakeTx(ctx context.Context, homeDirKey *dagger.Directory, passwordString string) (string, error) {
	baseKeyContainer := m.GenerateKey(ctx, homeDirKey, passwordString)

	pubKey, err := baseKeyContainer.
		WithEntrypoint([]string{"sh"}).
		WithExec([]string{"gnokey", "list", "-home=/gnohome"}).
		Stdout(ctx)

	if err != nil {
		return "", err
	}
	pubKey = m.parsePubAddr(pubKey)

	destMountDir := fmt.Sprintf("/gnopackages/%s", RealmName)

	return baseKeyContainer.
		WithServiceBinding("gno", m.RunGnolandNode(pubKey)).
		WithDirectory(destMountDir, m.loadGnoPackage(RealmName)).
		WithExec([]string{"maketx", "addpkg", "-home=/gnohome", "-insecure-password-stdin", "-chainid", ChainId,
			"-gas-fee", "1000000ugnot", "-gas-wanted", "3000000", "-max-deposit", "100000000ugnot",
			"-remote", "gno:26657",
			"-pkgdir", destMountDir, "-pkgpath", fmt.Sprintf("gno.land/%s", RealmName), KeyId},
			dagger.ContainerWithExecOpts{
				UseEntrypoint: true,
				Stdin:         fmt.Sprintf("%[1]s\n%[1]s\n", passwordString),
			}).
		Stdout(ctx)
}

// Loads a gno.land example from official examples folder
func (m *Gnokey) loadGnoPackage(packageName string) *dagger.Directory {
	repoDir := dag.Gnocloner().CloneMaster()
	return repoDir.Directory(fmt.Sprintf("./examples/gno.land/%s", packageName))
}

// Parse public address of gnokey
func (m *Gnokey) parsePubAddr(keyline string) string {
	start := strings.Index(keyline, "addr: ") + len("addr: ")
	end := strings.Index(keyline[start:], " ") + start
	return keyline[start:end]
}

// Run Gnoland chain
func (m *Gnokey) RunGnolandNode(
	// +optional
	publicKey string,
) *dagger.Service {
	// use entrypoint
	execOpts := dagger.ContainerWithExecOpts{
		UseEntrypoint: true,
	}

	// Add current account to genesis
	ctr := dag.Container().
		From("ghcr.io/gnolang/gno/gnoland:master")

	if publicKey != "" {
		ctr.
			WithExec([]string{"sh", "-c", fmt.Sprintf("echo %s=10000000000ugnot >> /gnoroot/gno.land/genesis/genesis_balances.txt", publicKey)})
	}

	gnolandSvc := ctr.
		// invalidate cache
		WithEnvVariable("CACHE_BUSTER", time.Now().Format(time.RFC3339Nano)).
		WithExec([]string{"config", "init"}, execOpts).
		WithExec([]string{"config", "set", "rpc.laddr", "tcp://0.0.0.0:26657"}, execOpts).
		WithExposedPort(26657).
		AsService(dagger.ContainerAsServiceOpts{
			Args:          []string{"start", "--lazy", "--skip-genesis-sig-verification", "--log-level", "info"},
			UseEntrypoint: execOpts.UseEntrypoint,
		})

	// wait for gnoland RPC service to be ready - can be useless
	code, _ := dag.Container().
		From("alpine:3").
		WithServiceBinding("gno", gnolandSvc).
		WithExec([]string{"apk", "add", "curl"}).
		WithExec([]string{"curl", "-fsS", "--retry", "5", "--retry-delay", "20", "--retry-all-errors", "http://gno:26657/status?height_gte=1"}).
		ExitCode(context.Background())

	if code != 0 {
		return nil
	}

	return gnolandSvc
}
