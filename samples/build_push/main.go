// A generated module for BuildPush functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"dagger/build-push/internal/dagger"
	"fmt"
)

type BuildPush struct{}

// Build the image
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
