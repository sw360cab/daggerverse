package main

import (
	"context"
	"dagger/k-3-s/internal/dagger"
	"fmt"
	"strings"
)

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

func getNodeAddess(ctx context.Context, nodeName string, gnoSecrets *dagger.Directory) string {
	// TODO: replace with node id from gnogenesis
	nodeKey, _ := dag.Container().From(GnolandBinary).
		WithDirectory("/gnoroot/gnoland-data/secrets", gnoSecrets).
		WithExec(strings.Split("secrets get node_id.id -raw", " ")).Stdout(ctx)

	return fmt.Sprintf("%s@%s:%s", nodeKey, nodeName, P2pPort)
}
