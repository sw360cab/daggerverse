// A generated module for K3S functions
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
	"dagger/k-3-s/internal/dagger"
	"strings"
	"time"
)

const NodeName string = "cluster.node"

type GnoK3s struct{}

// Starts a k3s server and deploys the Gnoland Helm chart by Core Team
func (m *GnoK3s) SpinCluster(
	ctx context.Context,
	// helm-related data folder
	helmDataFolder *dagger.Directory,
	// helm template folders
	helmTemplateFolder *dagger.Directory,
) (string, error) {
	k3s := dag.K3S("gnoland-test-cluster")
	kServer := k3s.Server()
	kServer, err := kServer.Start(ctx)
	if err != nil {
		return "", err
	}

	defaultFileOwner := dagger.ContainerWithFileOpts{Owner: "1001"}
	defaultDirOwner := dagger.ContainerWithDirectoryOpts{Owner: "1001"}

	// Secrets dir
	gnoSecretsDir := m.generateSecrets()

	// Genesis file
	genesisFile := m.generateGenesis(NodeName, gnoSecretsDir)

	initContainer := dag.Container().From("alpine/helm").
		WithoutEntrypoint().
		WithExec([]string{"apk", "add", "kubectl"}).
		// WithExec([]string{"apk", "add", "kustomize"}).
		WithEnvVariable("KUBECONFIG", "/.kube/config").
		WithFile("/.kube/config", k3s.Config(), defaultFileOwner).
		WithUser("1001").
		WithEnvVariable("CACHE_BUSTER", time.Now().Format(time.RFC3339Nano))

	return initContainer.
		WithDirectory("/opt/data/gno-secrets", gnoSecretsDir, defaultDirOwner). // Gnoland secrets and config
		WithFile("/opt/data/config/config.toml", helmDataFolder.File("config/config.toml"), defaultFileOwner).
		WithDirectory("/opt/data/genesis-server", helmDataFolder.Directory("genesis-server"), defaultDirOwner).
		WithFile("/opt/data/genesis.json", genesisFile, defaultFileOwner).
		WithFile("/opt/data/kustomization.yaml", helmDataFolder.File("kustomization.yaml"), defaultFileOwner).
		WithDirectory("/opt/data/helm", helmTemplateFolder, defaultDirOwner). // Helm template for Validator
		WithFile("/opt/data/template-values.yaml", helmDataFolder.File("template-values.yaml"), defaultFileOwner).
		WithWorkdir("/opt/data").
		WithExec([]string{"kubectl", "create", "ns", "gno"}).
		WithExec([]string{"kubectl", "create", "sa", "default", "-n", "gno"}).
		WithExec([]string{"kubectl", "apply", "-k", "."}).
		WithExec([]string{"kubectl", "apply", "-k", "genesis-server/"}).
		WithExec([]string{"kubectl", "wait", "--for=condition=ready", "--timeout=30s", "pod", "-l", "app=genesis-file-server", "-n", "gno"}).
		WithExec([]string{"kubectl", "cp", "/opt/data/genesis.json", "gno/genesis-file-server:/usr/share/nginx/html/genesis.json"}).
		Terminal().
		WithExec(strings.Split("helm install val-00 /opt/data/helm --values /opt/data/template-values.yaml --set global.genesisUrl=http://genesis-svc/genesis.json", " ")).
		WithExec([]string{"kubectl", "get", "pod", "-A"}).
		Stdout(ctx)
}

// Generates secrets using gnoland master
func (m *GnoK3s) generateSecrets() *dagger.Directory {
	return dag.Container().
		From("ghcr.io/gnolang/gno/gnoland:master").
		WithExec([]string{"secrets", "init"}, dagger.ContainerWithExecOpts{
			UseEntrypoint: true,
		}).
		Directory("/gnoroot/gnoland-data/secrets")
}

// Generates secrets using gnoland master
func (m *GnoK3s) generateGenesis(
	nodeName string,
	secretsFolder *dagger.Directory) *dagger.File {
	return dag.Gnogenesis().AddValidatorNode(
		nodeName,
		dag.Gnogenesis().Generate(),
		secretsFolder,
	)
}
