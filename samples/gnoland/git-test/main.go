// Download code from Git and run Tests
package main

import (
	"context"
	"dagger/gnoland/internal/dagger"
)

type Gnoland struct{}

type Locator string
type GitGno struct {
	Locator Locator
	Ref     string
}

const (
	Branch  Locator = "BRANCH"
	Tag     Locator = "TAG"
	Commit  Locator = "COMMIT"
	GnoRepo string  = "https://github.com/gnolang/gno.git"
)

// Clones git repository into a dir
func (m *Gnoland) clone(gitGno GitGno) *dagger.Directory {
	r := dag.Git(GnoRepo)
	var d *dagger.Directory

	ref := gitGno.Ref
	if ref == "" {
		ref = "master"
	}

	switch gitGno.Locator {
	case Tag:
		d = r.Tag(ref).Tree()
	case Commit:
		d = r.Commit(ref).Tree()
	case Branch:
	default:
		d = r.Branch(ref).Tree()
	}
	return d
}

// Clones MASTER git repository into a dir
// FIXME: needed in external module because custom type export does not work as expected
func (m *Gnoland) Clone() *dagger.Directory {
	return dag.Git(GnoRepo).Branch("master").Tree()
}

// Debugs using terminal
func (m *Gnoland) GitCodeBase(gitGno GitGno) *dagger.Container {
	return dag.Container().
		From("golang:1.22-alpine").
		WithDirectory("/src", m.Clone()).
		WithWorkdir("/src").
		WithExec([]string{"go", "test", "-v", "-count=1", "./gnovm/pkg/gnofmt"})
}

// Runs a test within dir holding the cloned repository
func (m *Gnoland) GitCodeTest(
	ctx context.Context,
	// +optional
	locator Locator,
	// +optional
	ref string) (string, error) {

	return m.GitCodeBase(GitGno{
		Locator: locator,
		Ref:     ref,
	}).Stdout(ctx)
}

// Debugs using terminal
func (m *Gnoland) GitCodeTestDebug(
	ctx context.Context,
	// +optional
	locator Locator,
	// +optional
	ref string) *dagger.Container {

	return m.GitCodeBase(GitGno{
		Locator: locator,
		Ref:     ref,
	}).Terminal()
}
