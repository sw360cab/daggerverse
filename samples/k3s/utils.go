package main

import (
	"context"
	"dagger/k-3-s/internal/dagger"
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
