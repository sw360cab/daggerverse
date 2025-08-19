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
	"fmt"
	"slices"
	"strings"
	"time"
)

const (
	NodeName    string = "cluster.node"
	K3sKubePort int    = 6443
)

type GnoK3s struct {
	initContainer  *dagger.Container
	kubeRepoFolder *dagger.Directory
	k3sEndpoint    string
}

var (
	defaultFileOwner = dagger.ContainerWithFileOpts{Owner: "1001"}
	defaultDirOwner  = dagger.ContainerWithDirectoryOpts{Owner: "1001"}
	validators       = map[string]validatorNode{}
)

// Starts a k3s server and deploys the Gnoland Helm chart by Core Team
func (m *GnoK3s) SpinCluster(
	ctx context.Context,
	// helm-related data folder
	helmDataFolder *dagger.Directory,
	// GitHub API token
	repoToken *dagger.Secret,
	// +optional
	// +default=betanet
	gitBranch string,
	// validators nodes
	// +optional
	// +default=2
	valCounter int,
	// sentry nodes
	// +optional
	// +default=0
	sentryCounter int,
	// validator-sentry ratio - how many validators behind a sentry
	// +optional
	// +default=2
	sentryRatio int,
) (int, error) {
	// initialize K3s cluster
	k3s := dag.K3S("gnoland-test-cluster")
	kServer := k3s.Server()
	_, err := kServer.Start(ctx)
	if err != nil {
		return -1, err
	}
	m.k3sEndpoint, _ = kServer.Endpoint(ctx, dagger.ServiceEndpointOpts{Port: K3sKubePort})

	// from Git Repo -> Kubernetes folder
	m.kubeRepoFolder = dag.
		Git("github.com/sw360cab/infrastructure", dagger.GitOpts{
			HTTPAuthUsername: "sw360cab",
			HTTPAuthToken:    repoToken,
		}).
		Branch(gitBranch).
		Tree().
		Directory("k8s")

	// Helm template dir
	helmTemplateFolder := m.kubeRepoFolder.Directory("core/helm")

	// generate basic genesis
	genesisFile := dag.Gnogenesis().Generate()

	// generate secrets for validator and add them to genesis
	for i := range valCounter {
		nodeName := fmt.Sprintf("gnocore-val-%02d", i+1)
		// Secrets dir
		gnoSecretsDir := m.generateSecrets()
		// Genesis file
		genesisFile = m.generateGenesis(nodeName, gnoSecretsDir)
		validators[nodeName] = validatorNode{
			name:          nodeName,
			secretsFolder: gnoSecretsDir,
		}
	}

	// generate RPC node - not added to genesis
	rpcNode := validatorNode{
		name:            fmt.Sprintf("gnocore-rpc-%02d", 1),
		secretsFolder:   m.generateSecrets(),
		configOverrides: RpcHelmValues,
	}

	// initalize cluster env
	m.initContainer = dag.Container().From("alpine/helm").
		WithoutEntrypoint().
		WithExec([]string{"apk", "add", "kubectl"}).
		// WithExec([]string{"apk", "add", "kustomize"}).
		WithEnvVariable("KUBECONFIG", "/.kube/config").
		WithFile("/.kube/config", k3s.Config(), defaultFileOwner).
		WithUser("1001").
		WithEnvVariable("CACHE_BUSTER", time.Now().Format(time.RFC3339Nano)).
		WithExec([]string{"kubectl", "create", "ns", "gno"})

	// TODO: sentry/validator config

	// bootstrap genesis server and helm files
	m.initContainer = m.initContainer.
		WithDirectory("/opt/data/genesis-server", helmDataFolder.Directory("genesis-server"), defaultDirOwner).
		WithFile("/opt/data/genesis.json", genesisFile, defaultFileOwner).
		WithFile("/opt/data/kustomization.yaml", helmDataFolder.File("kustomization.yaml"), defaultFileOwner).
		WithDirectory("/opt/data/helm", helmTemplateFolder, defaultDirOwner). // Helm template for Validator
		WithFile("/opt/data/template-values.yaml", helmDataFolder.File("template-values.yaml"), defaultFileOwner).
		WithWorkdir("/opt/data").
		WithExec([]string{"kubectl", "apply", "-k", "genesis-server/"}).
		WithExec([]string{"kubectl", "wait", "--for=condition=ready", "--timeout=30s", "pod", "-l", "app=genesis-file-server", "-n", "gno"}).
		WithExec([]string{"kubectl", "cp", "/opt/data/genesis.json", "gno/genesis-file-server:/usr/share/nginx/html/genesis.json"})

	// Spin validator nodes
	for valName, valNode := range validators {
		m.spinValidatorNode(valName, valNode, helmDataFolder)
	}

	// Spin RPC node
	m.spinValidatorNode(rpcNode.name, rpcNode, helmDataFolder)

	// Test RPC connection
	exitCode, err := m.testGnoservice(ctx, m.initContainer, rpcService.name, rpcService.port, rpcService.testPath)
	if err != nil {
		return exitCode, err
	}
	// TODO: Test blocks are produced

	// launch collateral services
	for _, svcValues := range gnoServices {
		// spin gno service
		svcContainer := m.spinGnoservice(ctx, svcValues.name, svcValues.deployDir)
		// test gno service
		exitCode, err = m.testGnoservice(ctx, svcContainer, svcValues.name, svcValues.port, svcValues.testPath)
		if err != nil {
			break
		}
	}
	return exitCode, err
}

// Spins a generic node that can be either a validator, sentry or rpc node
func (m *GnoK3s) spinValidatorNode(valName string, valNode validatorNode, helmDataFolder *dagger.Directory) *dagger.Container {
	homeFolder := fmt.Sprintf("/opt/data/%s", valName)
	return m.initContainer.
		// Gnoland secrets and config
		WithFile(fmt.Sprintf("%s/config/config.toml", homeFolder), helmDataFolder.File("config/config.toml"), defaultFileOwner).
		WithDirectory(fmt.Sprintf("%s/gno-secrets", homeFolder), valNode.secretsFolder, defaultDirOwner).
		WithExec([]string{"sh", "-c", fmt.Sprintf("sed -e 's/gnocore-val-01/%s/' /opt/data/kustomization.yaml > %s/kustomization.yaml", valName, homeFolder)}).
		WithExec([]string{"kubectl", "apply", "-k", homeFolder}).
		// Helm values and template
		WithExec([]string{"sh", "-c", fmt.Sprintf("sed -e 's/gnocore-val-01/%s/' /opt/data/template-values.yaml > %s/values.yaml", valName, homeFolder)}).
		WithExec(slices.Concat([]string{"helm", "install", valName, "/opt/data/helm", "--values", fmt.Sprintf("%s/values.yaml", homeFolder),
			"--set", "global.genesisUrl=http://genesis-svc/genesis.json"}, valNode.GetOverridesHelm())).
		WithExec([]string{"kubectl", "wait", "--for=condition=ready", "--timeout=60s", "pod", "-l", fmt.Sprintf("gno.name=%s", valName), "-n", "gno"})
}

// Spins up a Gno service which is not directly a validator node one
func (m *GnoK3s) spinGnoservice(
	ctx context.Context,
	serviceName string,
	serviceDirname string,
) *dagger.Container {
	// Gnoweb
	k8sYamlFiles := m.kubeRepoFolder.Directory(serviceDirname).Filter(dagger.DirectoryFilterOpts{
		Include: []string{"*/*yaml"},
		Exclude: []string{"ingress/*"},
	})
	filterdEntries, _ := k8sYamlFiles.Entries(ctx)

	gnoserviceContainer := m.initContainer
	var kubectlFlag string
	filePaths := getFiles(ctx, k8sYamlFiles, filterdEntries)

	// deploy resources
	for _, path := range filePaths {
		deployPath := path
		if strings.Contains(path, "kustomization.yaml") {
			kubectlFlag = "-k"
			deployPath = strings.ReplaceAll(path, "kustomization.yaml", "")
		} else {
			kubectlFlag = "-f"
		}
		gnoserviceContainer = gnoserviceContainer.
			WithFile("/opt/data/"+path, k8sYamlFiles.File(path), defaultFileOwner).
			WithWorkdir("/opt/data").
			WithExec([]string{"kubectl", "apply", kubectlFlag, deployPath})
	}
	// path service to make it testable
	return gnoserviceContainer.WithExec([]string{"kubectl", "patch",
		"service", serviceName,
		"-n", "gno",
		"-p", "{\"spec\":{\"type\":\"LoadBalancer\"}}"})
}

// test a gno service at the Given Service Url obtained mixing
// - the service name
// - the service port randomly assigned by LoadBalancer, obtained through servicePort
// - the url path to test
func (m *GnoK3s) testGnoservice(
	ctx context.Context,
	testableContainer *dagger.Container,
	serviceName string,
	servicePort int,
	testPath string) (int, error) {
	if testPath == "" {
		testPath = "/"
	}
	svcPort, _ := testableContainer.
		WithExec(strings.Split("kubectl get svc -n gno "+
			serviceName+
			" -o jsonpath='{.spec.ports[?(@.port=="+
			fmt.Sprintf("%d", servicePort)+
			")].nodePort}'", " ")).
		Stdout(ctx)
	svcPort = strings.ReplaceAll(svcPort, "'", "")
	svcUrl := strings.ReplaceAll(m.k3sEndpoint, fmt.Sprintf("%d", K3sKubePort), svcPort)

	return testableContainer.
		WithExec([]string{"curl", "--retry", "5", "--retry-delay", "5", "--retry-all-errors", fmt.Sprintf("http://%s%s", svcUrl, testPath)}).
		ExitCode(ctx)
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
