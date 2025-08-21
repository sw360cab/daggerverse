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

type networkNode struct {
	name            string
	nodeAddress     string // node p2p address
	secretsFolder   *dagger.Directory
	configOverrides map[string]string
}

// Returns all the ovverides config params ready for helm ... -set key=value
func (v *networkNode) GetOverridesHelm() (configItems []string) {
	for key, val := range v.configOverrides {
		configItems = append(configItems, "--set", fmt.Sprintf("%s=%s", key, val))
	}
	return
}
