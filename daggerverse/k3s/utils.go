package main

import (
	"context"
	"dagger/k-3-s/internal/dagger"
	"fmt"
)

// Retrieves given files as entries of names/paths from a specific Directory
func getFiles(ctx context.Context, rootDir *dagger.Directory, entries []string) []string {
	var items []string
	for _, dirPath := range entries {
		files, _ := rootDir.Directory(dirPath).Entries(ctx)
		for _, filePath := range files {
			items = append(items, dirPath+filePath)
		}
	}
	return items
}

// Composes node address from id + hostname + port
func getNodeAddress(ctx context.Context, nodeName string, gnoSecrets *dagger.Directory) string {
	nodeKey, _ := dag.Gnogenesis().GetNodeID(ctx, gnoSecrets)
	return fmt.Sprintf("%s@%s%s:%s", nodeKey, nodeName, SvcSuffix, P2pPort)
}
