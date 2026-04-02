package main

import (
	"os"
	"path/filepath"
)

// buildTree creates a node for the given path.
// If shallow is true, directory children are not loaded (lazy loading).
func buildTree(rootPath string) fileNode {
	return buildTreeShallow(rootPath, true)
}

// buildTreeShallow creates a node with optional shallow loading.
// When shallow is true, directory children are not loaded until expanded.
func buildTreeShallow(rootPath string, shallow bool) fileNode {
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
			if shallow {
				// Only create placeholder nodes without loading their children
				childInfo, err := entry.Info()
				if err != nil {
					continue
				}
				node.children = append(node.children, fileNode{
					name:  entry.Name(),
					path:  childPath,
					isDir: childInfo.IsDir(),
				})
			} else {
				// Recursively load all children
				node.children = append(node.children, buildTreeShallow(childPath, false))
			}
		}
		node.loaded = true
	}

	return node
}

// loadChildren loads children for a directory node if not already loaded.
// Returns true if children were loaded, false if already loaded.
func (n *fileNode) loadChildren() bool {
	if n.loaded || !n.isDir {
		return false
	}

	entries, err := os.ReadDir(n.path)
	if err != nil {
		return true // Mark as loaded even on error to prevent retries
	}

	for _, entry := range entries {
		childPath := filepath.Join(n.path, entry.Name())
		childInfo, err := entry.Info()
		if err != nil {
			continue
		}
		n.children = append(n.children, fileNode{
			name:  entry.Name(),
			path:  childPath,
			isDir: childInfo.IsDir(),
		})
	}
	n.loaded = true
	return true
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