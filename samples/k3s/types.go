package main

import (
	"dagger/k-3-s/internal/dagger"
	"fmt"
)

type gnoService struct {
	name      string
	deployDir string // folder within project
	port      int    // service internl port
	testPath  string // endpoint path to be used while testing readiness
}

// type validatorConfig struct {
// 	bootnodes    []string // p2p.seeds
// 	privatePeers []string // p2p.private_peer_ids -> only for sentries
// }

type validatorNode struct {
	name          string
	secretsFolder *dagger.Directory
	// p2pOverrides    *validatorConfig
	configOverrides map[string]string
}

// Returns all the ovverides config params ready for helm ... -set key=value
func (v *validatorNode) GetOverridesHelm() (configItems []string) {
	for key, val := range v.configOverrides {
		configItems = append(configItems, "--set", fmt.Sprintf("%s=%s", key, val))
	}
	return
}
