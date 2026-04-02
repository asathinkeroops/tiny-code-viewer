package main

import (
	"os"
	"path/filepath"
)

func buildTree(rootPath string) fileNode {
	info, err := os.Stat(rootPath)
	if err != nil {
		return fileNode{name: rootPath, path: rootPath}
	}

	node := fileNode{
		name:  info.Name(),
		path:  rootPath,
		isDir: info.IsDir(),
	}

	if node.isDir {
		entries, err := os.ReadDir(rootPath)
		if err != nil {
			return node
		}
		for _, entry := range entries {
			childPath := filepath.Join(rootPath, entry.Name())
			node.children = append(node.children, buildTree(childPath))
		}
	}

	return node
}

func flattenTree(node fileNode, expanded map[string]bool, depth int) []itemInfo {
	var result []itemInfo
	result = append(result, itemInfo{path: node.path, isDir: node.isDir, depth: depth})

	if node.isDir && expanded[node.path] {
		for _, child := range node.children {
			result = append(result, flattenTree(child, expanded, depth+1)...)
		}
	}

	return result
}

func (m *model) getVisibleItems() []itemInfo {
	return flattenTree(m.root, m.expanded, 0)
}