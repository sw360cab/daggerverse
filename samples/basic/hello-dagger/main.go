package main

import (
	"context"
	"dagger/gno-dagger/internal/dagger"
)

type GnoDagger struct{}

// Returns a container that echoes whatever string argument is provided
func (m *GnoDagger) ContainerEcho(stringArg string) *dagger.Container {
	return dag.Container().From("alpine:latest").WithExec([]string{"echo", stringArg})
}

// Returns current platform.
// - forcing linux/amd64 -> x86_64
// - on Mac Silicon 		 -> aarch64
func (m *GnoDagger) Platform(ctx context.Context) (string, error) {
	return dag.Container(dagger.ContainerOpts{Platform: "linux/amd64"}).
		From("alpine:latest").
		WithExec([]string{"uname", "-m"}).
		Stdout(ctx)
}
