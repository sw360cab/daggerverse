package main

import (
	"context"
	"dagger/kind-dagger/internal/dagger"
	"strings"
)

type KindDagger struct{}

func (m *KindDagger) spinKind(ctx context.Context, socket *dagger.Socket) *dagger.KindCluster {
	kindCluster := dag.
		Kind(socket).
		Cluster(dagger.KindClusterOpts{
			Name: "dagger0",
		})

	_, err := kindCluster.Create(ctx)
	if err != nil {
		return nil
	}
	return kindCluster
}

// run using
// dagger call -i run-dagger-helm --dagger-version 0.18.14 --dagger-cli-helm ../dagger-helm/ --socket=/var/run/docker.sock
func (m *KindDagger) RunDaggerHelm(
	ctx context.Context,
	socket *dagger.Socket,
	daggerCliHelm *dagger.Directory,
	// +optional
	// +default=latest
	daggerVersion string,

) (int, error) {
	kindCluster := m.spinKind(ctx, socket)
	defer kindCluster.Delete(ctx)

	daggerEngineContainer := dag.DaggerHelm().
		InstallDaggerHelm(daggerVersion, kindCluster.Kubeconfig(
			dagger.KindClusterKubeconfigOpts{
				Internal: true,
			},
		))

	daggerCliHelmContainer := dag.DaggerHelm().RunDaggerCliHelm(
		daggerVersion,
		daggerCliHelm,
		daggerEngineContainer,
	)

	return daggerCliHelmContainer.
		WithExec(append([]string{"kubectl", "exec", "dagger-cli-pod", "-n", "dagger", "--"},
			strings.Split("dagger call -m github.com/shykes/daggerverse/hello hello --giant --name daggernauts", " ")...)).
		Terminal().
		ExitCode(ctx)
}
