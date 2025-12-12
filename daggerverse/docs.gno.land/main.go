package main

import (
	"context"
	"dagger/docs-gno-land/internal/dagger"
)

type DocsGnoLand struct{}

// NOTE: using Platform: "linux/amd64" for base image container to avoid problem on building constraints
func (m *DocsGnoLand) buildDoc() *dagger.Container {
	return dag.Container(dagger.ContainerOpts{Platform: "linux/amd64"}).
		From("node:20.18").
		WithMountedDirectory("/mnt", dag.Git("https://github.com/gnolang/docs.gno.land/").Head().Tree()).
		WithWorkdir("/mnt/docusaurus").
		WithExec([]string{"yarn", "run", "download-docs"}).
		WithExec([]string{"yarn", "install", "--verbose"}).
		WithExec([]string{"yarn", "build"})
}

// Builds the doc and return output
func (m *DocsGnoLand) BuildDoc(ctx context.Context) (string, error) {
	return m.buildDoc().
		Stdout(ctx)
}

// Preview publish to Netlify
func (m *DocsGnoLand) PublishDoc(ctx context.Context, authToken *dagger.Secret, siteId *dagger.Secret) (string, error) {
	netlify := m.buildDoc().
		WithSecretVariable("NETLIFY_AUTH_TOKEN", authToken).
		WithSecretVariable("NETLIFY_SITE_ID", siteId).
		WithExec([]string{"npm", "install", "-g", "netlify-cli"})

	return netlify.
		WithExec([]string{"netlify", "deploy", "--prod", "false", "--debug", "true"}).
		Stdout(ctx)
}
