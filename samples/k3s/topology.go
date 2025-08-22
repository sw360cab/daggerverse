package main

import (
	"context"
	"fmt"
	"maps"
	"strings"
)

const MaxValidatorCount = 10

// Vaidates topology
func IsValidTopology(valCounter, sentryCounter, sentryRatio int) error {
	if valCounter == 2 {
		return fmt.Errorf("invalid valCounter: %d impossible to reach 2/3 majority", valCounter)
	}

	if valCounter <= 0 || valCounter >= MaxValidatorCount {
		return fmt.Errorf("invalid valCounter: %d (must be between 1 and %d)", valCounter, MaxValidatorCount)
	}

	if sentryRatio < 1 {
		return fmt.Errorf("invalid sentryRatio: %d (must be greater than 0)", sentryRatio)
	}

	if sentryCounter == 0 { // no more checks needed
		return nil
	}

	// Ratio validator
	minVals := (sentryCounter-1)*sentryRatio + 1
	maxVals := sentryCounter * sentryRatio
	if valCounter < minVals || valCounter > maxVals {
		return fmt.Errorf(
			"invalid topology: with %d sentries and ratio %d, validators must be between %d and %d",
			sentryCounter, sentryRatio, minVals, maxVals,
		)
	}

	return nil
}

// Generates secrets for validator and add them to genesis
func (m *GnoK3s) setupValidatorNodes(ctx context.Context, valCounter int) []networkNode {
	validators := []networkNode{}
	for i := range valCounter {
		nodeName := fmt.Sprintf("gnocore-val-%02d", i+1)
		// Secrets dir
		gnoSecretsDir := m.generateSecrets()
		// Genesis file
		m.genesisFile = m.generateGenesis(nodeName, gnoSecretsDir)
		validators = append(validators, networkNode{
			name:          nodeName,
			nodeAddress:   getNodeAddress(ctx, nodeName, gnoSecretsDir),
			secretsFolder: gnoSecretsDir,
		})
	}
	return validators
}

// Generate secrets for sentries
// then connects validators to sentry in configuration according to `sentryRatio` and topology
// Eventually after all is set adjusts configurations of sentries and validators
// - For each sentry add all the other sentries as seed nodeby updating configration / - For each validator add mutual sentry node by updating configration
func (m *GnoK3s) setupSentryNodes(ctx context.Context, validators []networkNode, sentryCounter, sentryRatio int) ([]networkNode, error) {
	sentries := []networkNode{}
	for i := range sentryCounter {
		if sentryCounter == 0 {
			break
		}
		nodeName := fmt.Sprintf("gnocore-sentry-%02d", i+1)
		// Secrets dir
		gnoSecretsDir := m.generateSecrets()

		// get topology addresses
		nodeAddress := getNodeAddress(ctx, nodeName, gnoSecretsDir)
		// this sentry
		peers := []string{nodeAddress}
		privateIds := []string{}
		// peer validators
		for j := (i * sentryRatio); j < (i+1)*sentryRatio; j++ {
			if j > len(validators)-1 { // no more validators to be added
				break
			}
			validatorPrivateId, err := dag.Gnogenesis().GetNodeID(ctx, gnoSecretsDir)
			if err != nil {
				return nil, err
			}
			privateIds = append(privateIds, validatorPrivateId)
			peers = append(peers, validators[j].nodeAddress)
		}

		// sentry overrides
		overrides := maps.Clone(SentryHelmValues)
		overrides[P2pPeersHelmKey] = strings.Join(peers, "\\,")
		overrides[P2pSeedHelmKey] = strings.Join(peers, "\\,") // same as peers
		overrides[P2pPrivateIdsHelmKey] = strings.Join(privateIds, "\\,")

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
				sentryNode.configOverrides[P2pPeersHelmKey] += "\\," + otherSentryNode.nodeAddress
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
		peersStr := strings.Join(peers, "\\,")
		validatorNode.configOverrides = map[string]string{}
		validatorNode.configOverrides[P2pPeersHelmKey] = peersStr
		validatorNode.configOverrides[P2pSeedHelmKey] = peersStr
	}
	return sentries, nil
}
