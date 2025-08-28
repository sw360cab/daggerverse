package main

import (
	"context"
	"dagger/gnokey/internal/dagger"
	"fmt"
	"strings"
)

type Gnokey struct{}

const (
	ChainId   string = "dev"
	KeyId     string = "key0"
	RealmName string = "r/demo/counter"
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

	destMountDir := fmt.Sprintf("/gnopackages/%s", RealmName)

	return baseKeyContainer.
		WithServiceBinding("gno", m.RunGnolandValidator(pubKey)).
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

func (m *Gnokey) loadGnoPackage(packageName string) *dagger.Directory {
	repoDir := dag.Gitcloner().CloneMaster()
	return repoDir.Directory(fmt.Sprintf("./examples/gno.land/%s", packageName))
}

// Run Gnoland chain
func (m *Gnokey) RunGnolandValidator(publicKey string) *dagger.Service {
	// use entrypoint
	execOpts := dagger.ContainerWithExecOpts{
		UseEntrypoint: true,
	}

	// Add current account to genesis
	ctr := dag.Container().
		From("ghcr.io/gnolang/gno/gnoland:chain-test5.0")

	if publicKey != "" {
		ctr.
			WithExec([]string{"sh", "-c", fmt.Sprintf("echo %s=10000000000ugnot >> /gnoroot/gno.land/genesis/genesis_balances.txt", publicKey)})
	}

	return ctr.
		WithExposedPort(26657).
		WithExec([]string{"config", "init"}, execOpts).
		WithExec([]string{"config", "set", "rpc.laddr", "tcp://0.0.0.0:26657"}, execOpts).
		AsService(dagger.ContainerAsServiceOpts{
			Args:          []string{"start", "--lazy", "--log-level", "info", "--chainid", ChainId},
			UseEntrypoint: execOpts.UseEntrypoint,
		})
}

func (m *Gnokey) parsePubAddr(keyline string) string {
	start := strings.Index(keyline, "addr: ") + len("addr: ")
	end := strings.Index(keyline[start:], " ") + start
	return keyline[start:end]
}
