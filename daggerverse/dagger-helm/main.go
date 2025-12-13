// Run a full cluster into K3s using repository files
package main

import (
	"context"
	"dagger/dagger-helm/internal/dagger"
	"fmt"
	"slices"
	"strings"
	"time"
)

const (
	ClusterName string = "daggerhelm.cluster.test"
	K3sKubePort int    = 6443
)

type DaggerHelmK3s struct{}

var (
	defaultFileOwner = dagger.ContainerWithFileOpts{Owner: "1001"}
	defaultDirOwner  = dagger.ContainerWithDirectoryOpts{Owner: "1001"}
)

func (m *DaggerHelmK3s) DaggerHelmK3s(
	ctx context.Context,
	daggerCliHelm *dagger.Directory,
	// +optional
	// +default=latest
	daggerVersion string,
) (int, error) {
	// initialize K3s cluster
	k3s := dag.K3S(ClusterName)
	kServer := k3s.Server()
	_, err := kServer.Start(ctx)
	if err != nil {
		return -1, err
	}

	daggerEngineContainer := m.InstallDaggerHelm(
		daggerVersion,
		k3s.Config(),
	)

	daggerCliHelmContainer := m.RunDaggerCliHelm(
		daggerVersion,
		daggerCliHelm,
		daggerEngineContainer,
	)

	return m.runDaggerHelloModule(ctx, daggerCliHelmContainer)
}

func (m *DaggerHelmK3s) InstallDaggerHelm(
	daggerVersion string,
	kubeConfig *dagger.File,
) *dagger.Container {
	// initalize cluster env
	initContainer := dag.Container().From("alpine/helm").
		WithoutEntrypoint().
		WithExec([]string{"apk", "add", "kubectl"}).
		WithEnvVariable("KUBECONFIG", "/.kube/config").
		WithFile("/.kube/config", kubeConfig, defaultFileOwner).
		WithUser("1001").
		WithEnvVariable("HOME", "/tmp").
		WithWorkdir("/tmp").
		WithEnvVariable("CACHE_BUSTER", time.Now().Format(time.RFC3339Nano))

	helmDaggerCmd := []string{"helm", "install", "dagger-engine", "oci://registry.dagger.io/dagger-helm",
		"--namespace", "dagger", "--create-namespace"}

	if daggerVersion != "latest" {
		helmDaggerCmd = slices.Concat(helmDaggerCmd, []string{"--version", daggerVersion})
	}

	// install Dagger Helm Chart
	return initContainer.
		// WithDirectory("/run/dagger-dagger-engine-dagger-helm", daggerCliHelm.Directory("./runner"), defaultDirOwner).
		// WithExec([]string{"chmod", "777", "/run/dagger-dagger-engine-dagger-helm"}).
		WithExec(helmDaggerCmd)
}

func (m *DaggerHelmK3s) RunDaggerCliHelm(
	daggerVersion string,
	daggerCliHelm *dagger.Directory,
	daggerEngineContainer *dagger.Container,
) *dagger.Container {
	// install DaggerCli Custom Helm Chart
	return daggerEngineContainer.
		WithDirectory("/opt/data/helm", daggerCliHelm.Directory("./helm"), defaultDirOwner).
		WithExec([]string{"kubectl", "create", "serviceaccount", "default", "-n", "dagger"}).
		WithExec([]string{"helm", "install", "dagger-cli", "/opt/data/helm",
			"--set", fmt.Sprintf("daggerVersion=%s", daggerVersion),
			"--namespace", "dagger"}).
		WithExec([]string{"kubectl", "wait", "--for=condition=ready", "--timeout=90s", "pod", "-l", "app=dagger-cli", "-n", "dagger"})
}

func (m *DaggerHelmK3s) runDaggerHelloModule(
	ctx context.Context,
	daggerCliContainer *dagger.Container,
) (int, error) {
	return daggerCliContainer.
		WithExec(append([]string{"kubectl", "exec", "dagger-cli-pod", "-n", "dagger", "--"},
			strings.Split("dagger call -m github.com/shykes/daggerverse/hello hello --giant --name daggernauts", " ")...)).
		ExitCode(ctx)
}
