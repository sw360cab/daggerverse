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
	"maps"
	"slices"
	"strings"
	"time"
)

const (
	ClusterName string = "gnoland.cluster.test"
	K3sKubePort int    = 6443
)

type GnoK3s struct {
	initContainer  *dagger.Container
	kubeRepoFolder *dagger.Directory
	k3sEndpoint    string
	genesisFile    *dagger.File
}

var (
	defaultFileOwner = dagger.ContainerWithFileOpts{Owner: "1001"}
	defaultDirOwner  = dagger.ContainerWithDirectoryOpts{Owner: "1001"}
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
	// +default=1
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
	// init vars
	validators := []networkNode{}
	sentries := []networkNode{}

	// initialize K3s cluster
	k3s := dag.K3S(ClusterName)
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
	m.genesisFile = dag.Gnogenesis().Generate()

	// generate secrets for validator and add them to genesis
	for i := range valCounter {
		nodeName := fmt.Sprintf("gnocore-val-%02d", i+1)
		// Secrets dir
		gnoSecretsDir := m.generateSecrets()
		// Genesis file
		m.genesisFile = m.generateGenesis(nodeName, gnoSecretsDir)
		validators = append(validators, networkNode{
			name:          nodeName,
			nodeAddress:   getNodeAddess(ctx, nodeName, gnoSecretsDir),
			secretsFolder: gnoSecretsDir,
		})
	}

	// generate secrets for sentries
	// and cofigure validator address according to `sentryRatio`
	for i := range sentryCounter {
		if sentryCounter == 0 {
			break
		}
		nodeName := fmt.Sprintf("gnocore-sentry-%02d", i+1)
		// Secrets dir
		gnoSecretsDir := m.generateSecrets()

		// get topology addresses
		nodeAddress := getNodeAddess(ctx, nodeName, gnoSecretsDir)
		// this sentry
		peers := []string{nodeAddress}
		privateIds := []string{}
		// peer validators
		for j := (i * sentryRatio); j < (i+1)*sentryRatio; j++ {
			if j > len(validators)-1 { // no more validators to be added
				break
			}
			// validatorPrivateId, err := dag.Gnogenesis().GetNodeID(ctx, gnoSecretsDir)
			// if err != nil {
			// 	return -1, err
			// }
			// privateIds = append(privateIds, validatorPrivateId)
			peers = append(peers, validators[j].nodeAddress)
		}

		// sentry overrides
		overrides := maps.Clone(SentryHelmValues)
		overrides[P2pPeersYaml] = strings.Join(peers, ",")
		overrides[P2pSeedYaml] = strings.Join(peers, ",") // same as peers
		overrides[P2pPrivateIdsYaml] = strings.Join(privateIds, ",")

		sentries = append(sentries, networkNode{
			name:            nodeName,
			nodeAddress:     nodeAddress,
			secretsFolder:   gnoSecretsDir,
			configOverrides: overrides,
		})
	}

	// add other sentries to each sentry node config
	for index, sentryNode := range sentries {
		for otherIndex, otherSentryNode := range sentries {
			if otherIndex != index {
				sentryNode.configOverrides[P2pPeersYaml] += "," + otherSentryNode.nodeAddress
			}
		}
	}

	// configure validators P2P config overrides
	for index, validatorNode := range validators {
		if sentryCounter == 0 {
			break
		}
		peers := []string{sentries[index/sentryRatio].nodeAddress} // add sentry
		for otherIndex := (index / sentryRatio) * sentryRatio; otherIndex < ((index/sentryRatio)+1)*sentryRatio; otherIndex++ {
			if otherIndex != index {
				peers = append(peers, validatorNode.nodeAddress)
			}
		}
		peersStr := strings.Join(peers, ",")
		validatorNode.configOverrides = map[string]string{}
		validatorNode.configOverrides[P2pPeersYaml] = peersStr
		validatorNode.configOverrides[P2pSeedYaml] = peersStr
	}

	// generate RPC node - not added to genesis
	rpcNode := networkNode{
		name:            fmt.Sprintf("gnocore-rpc-%02d", 1),
		secretsFolder:   m.generateSecrets(),
		configOverrides: RpcHelmValues,
	}

	// initalize cluster env
	m.initContainer = dag.Container().From("alpine/helm").
		WithoutEntrypoint().
		WithExec([]string{"apk", "add", "kubectl"}).
		WithExec([]string{"apk", "add", "jq"}). // TODO: to be removed when /ready endpoint will be available for RPC
		WithEnvVariable("KUBECONFIG", "/.kube/config").
		WithFile("/.kube/config", k3s.Config(), defaultFileOwner).
		WithUser("1001").
		WithEnvVariable("CACHE_BUSTER", time.Now().Format(time.RFC3339Nano)).
		WithExec([]string{"kubectl", "create", "ns", "gno"})

	// bootstrap genesis server and helm files
	m.initContainer = m.initContainer.
		WithDirectory("/opt/data/genesis-server", helmDataFolder.Directory("genesis-server"), defaultDirOwner).
		WithFile("/opt/data/genesis.json", m.genesisFile, defaultFileOwner).
		WithFile("/opt/data/kustomization.yaml", helmDataFolder.File("kustomization.yaml"), defaultFileOwner).
		WithDirectory("/opt/data/helm", helmTemplateFolder, defaultDirOwner). // Helm template for Validator
		WithFile("/opt/data/template-values.yaml", helmDataFolder.File("template-values.yaml"), defaultFileOwner).
		WithWorkdir("/opt/data").
		WithExec([]string{"kubectl", "apply", "-k", "genesis-server/"}).
		WithExec([]string{"kubectl", "wait", "--for=condition=ready", "--timeout=30s", "pod", "-l", "app=genesis-file-server", "-n", "gno"}).
		WithExec([]string{"kubectl", "cp", "/opt/data/genesis.json", "gno/genesis-file-server:/usr/share/nginx/html/genesis.json"})

	// spin network nodes
	for _, netNode := range slices.Concat(validators, sentries) {
		m.initContainer = m.spinNetworkNode(netNode.name, netNode, helmDataFolder)
	}

	// spin RPC node
	m.initContainer = m.spinNetworkNode(rpcNode.name, rpcNode, helmDataFolder)
	//test RPC connection
	exitCode, err := m.testGnoservice(ctx, m.initContainer, rpcService.name, rpcService.port, rpcService.testPath)
	if err != nil {
		return exitCode, err
	}
	// TODO: test blocks are being produced
	_, err = m.GetSvcExposedEndpoint(ctx, m.initContainer, rpcService.name, rpcService.port)
	if err != nil {
		return -1, err
	}
	// exitCode, err = m.initContainer.
	// 	WithExec([]string{"sh", "-c", fmt.Sprintf("curl --retry 5 --retry-delay 5 --retry-all-errors -fsS %s/status | tee /dev/stderr | jq -r '.result.sync_info.latest_block_height') -ge 1", rpcUrl)}).
	// 	ExitCode(ctx)
	// if err != nil {
	// 	return exitCode, err
	// }

	// launch collateral services
	for _, svcValues := range gnoServices {
		// spin gno service
		svcContainer := m.spinGnoservice(ctx, svcValues.name, svcValues.deployDir)
		// test gno service
		_, err := m.testGnoservice(ctx, svcContainer, svcValues.name, svcValues.port, svcValues.testPath)
		if err != nil {
			break
		}
	}
	return exitCode, err
}

// Spins a network node that can be either a validator, sentry or rpc node
func (m *GnoK3s) spinNetworkNode(valName string, valNode networkNode, helmDataFolder *dagger.Directory) *dagger.Container {
	homeFolder := fmt.Sprintf("/opt/data/%s", valName)
	return m.initContainer.
		WithFile(fmt.Sprintf("%s/config/config.toml", homeFolder), helmDataFolder.File("config/config.toml"), defaultFileOwner).
		WithDirectory(fmt.Sprintf("%s/gno-secrets", homeFolder), valNode.secretsFolder, defaultDirOwner).
		WithExec([]string{"sh", "-c", fmt.Sprintf("sed -e 's/gnocore-val-01/%s/' /opt/data/kustomization.yaml > %s/kustomization.yaml", valName, homeFolder)}).
		WithExec([]string{"kubectl", "apply", "-k", homeFolder}).
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

// Gets exposed endpoint of load balanced svc
// by replacing svc target port with assigned cluster port
func (m *GnoK3s) GetSvcExposedEndpoint(
	ctx context.Context,
	testableContainer *dagger.Container,
	serviceName string,
	servicePort int) (string, error) {
	svcPort, err := testableContainer.
		WithExec(strings.Split("kubectl get svc -n gno "+
			serviceName+
			" -o jsonpath='{.spec.ports[?(@.port=="+
			fmt.Sprintf("%d", servicePort)+
			")].nodePort}'", " ")).
		Stdout(ctx)
	if err != nil {
		return "", err
	}
	svcPort = strings.ReplaceAll(svcPort, "'", "")
	return strings.ReplaceAll(m.k3sEndpoint, fmt.Sprintf("%d", K3sKubePort), svcPort), nil
}

// Tests a gno service at the Given Service Url obtained mixing
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
	svcUrl, err := m.GetSvcExposedEndpoint(ctx, testableContainer, serviceName, servicePort)
	if err != nil {
		return -1, err
	}

	return testableContainer.
		WithExec([]string{"curl", "--retry", "5", "--retry-delay", "5", "--retry-all-errors", fmt.Sprintf("http://%s%s", svcUrl, testPath)}).
		// Terminal().
		ExitCode(ctx)
}

// Generates secrets using gnoland master
func (m *GnoK3s) generateSecrets() *dagger.Directory {
	return dag.Container().
		From("ghcr.io/gnolang/gno/gnoland:master").
		// invalidate cache to replicate secret execution --> see https://docs.dagger.io/cookbook/#invalidate-cache
		WithEnvVariable("CACHEBUSTER", time.Now().String()).
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
		m.genesisFile,
		secretsFolder,
	)
}
