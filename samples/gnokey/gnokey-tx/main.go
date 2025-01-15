package main

import (
	"context"
	"dagger/gnokey/internal/dagger"
	"fmt"
	"strings"
)

type Gnokey struct{}

const (
	ChainId string = "dev"
	KeyId   string = "key0"
)

// Run Gnokey Container
func (m *Gnokey) BaseGnokey(homeDirKey *dagger.Directory) *dagger.Container {
	return dag.Container().
		From("ghcr.io/gnolang/gno/gnokey:master").
		WithMountedDirectory("/gnohome", homeDirKey)
}

// Generates a Gno key providing output
// TODO: remove
func (m *Gnokey) GenerateKeyOut(ctx context.Context, homeDirKey *dagger.Directory, passwordString string) (string, error) {
	ctd := m.GenerateKey(ctx, homeDirKey, passwordString)
	ctd.Directory("/gnohome").Export(ctx, "/tmp/")
	return ctd.Stdout(ctx)
}

// Generates a Gno key
func (m *Gnokey) GenerateKey(ctx context.Context, homeDirKey *dagger.Directory, passwordString string) *dagger.Container {
	return m.BaseGnokey(homeDirKey).
		WithExec([]string{"gnokey", "add", KeyId, "-home=/gnohome", "-insecure-password-stdin"},
			dagger.ContainerWithExecOpts{
				Stdin: fmt.Sprintf("%[1]s\n%[1]s\n", passwordString),
			})
}

// Perform a Tx using Gnokey on a local chain
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

	return baseKeyContainer.
		WithServiceBinding("gno", m.runGnolandValidator(pubKey, homeDirKey)).
		WithExec([]string{"maketx", "addpkg", "-home=/gnohome", "-insecure-password-stdin", "-chainid", ChainId,
			"-gas-fee", "1000000ugnot", "-gas-wanted", "3000000", "-deposit", "100000000ugnot",
			"-remote", "gno:26657",
			"-pkgdir", "/gnohome/gnopackage/counter", "-pkgpath", "gno.land/r/demo/counter", KeyId},
			dagger.ContainerWithExecOpts{
				UseEntrypoint: true,
				Stdin:         fmt.Sprintf("%[1]s\n%[1]s\n", passwordString),
			}).
		Stdout(ctx)
}

// Run Gnoland chain
func (m *Gnokey) runGnolandValidator(publicKey string, homeDirKey *dagger.Directory) *dagger.Service {
	// use entrypont
	execOpts := dagger.ContainerWithExecOpts{
		UseEntrypoint: true,
	}

	ctr := dag.Container().
		From("ghcr.io/gnolang/gno/gnoland:master").
		WithMountedDirectory("/tmp/appender", homeDirKey).
		WithExec([]string{"sh", "/tmp/appender/script.sh", fmt.Sprintf("%s=10000000000ugnot", publicKey)})
	// TODO: check these
	// WithExec([]string{"cp", "/gnoroot/gno.land/genesis/genesis_balances.txt", "/tmp/genesis_balances.txt"}).
	// WithExec([]string{"sh", "-c", "echo", fmt.Sprintf("%s=10000000000ugnot", publicKey), ">", "/tmp/genesis_balances.txt"})

	return ctr.
		From("ghcr.io/gnolang/gno/gnoland:master").
		WithExposedPort(26657).
		WithExec([]string{"config", "init"}, execOpts).
		WithExec([]string{"config", "set", "rpc.laddr", "tcp://0.0.0.0:26657"}, execOpts).
		WithExec([]string{"start", "--lazy", "--chainid", ChainId}, execOpts).
		AsService()
}

func (m *Gnokey) parsePubAddr(keyline string) string {
	start := strings.Index(keyline, "addr: ") + len("addr: ")
	end := strings.Index(keyline[start:], " ") + start
	return keyline[start:end]
}
