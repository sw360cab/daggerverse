package main

import (
	"context"
	"dagger/build-push/internal/dagger"
	"fmt"
)

type BuildPush struct{}

// Build a Gnoland image. `buildContext` is the folder where gno.land monorepo is stored locally
func (m *BuildPush) Build(ctx context.Context, buildContext *dagger.Directory) *dagger.Container {
	return buildContext.DockerBuild(dagger.DirectoryDockerBuildOpts{
		Target: "portalloopd",
	})
}

// Take the built container and push it
func (m *BuildPush) BuildAndPush(
	ctx context.Context,
	buildContext *dagger.Directory,
	registryEndpoint,
	imageName,
	// +optional
	tag,
	username string,
	token *dagger.Secret) error {
	if tag == "" {
		tag = "latest"
	}

	_, err := m.Build(ctx, buildContext).
		WithRegistryAuth(registryEndpoint, username, token).
		Publish(ctx, fmt.Sprintf("%s/%s:%s", registryEndpoint, imageName, tag))
	return err
}
