package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbletea"
)

func initialModel() *model {
	startPath := "."
	if len(os.Args) > 1 {
		startPath = os.Args[1]
	}
	absPath, _ := filepath.Abs(startPath)

	m := &model{
		root:        buildTree(absPath),
		rootPath:    absPath,
		expanded:    make(map[string]bool),
		focusLeft:   true,
		cursor:      1, // Skip root node, start at first child
		watcher:     initialWatcher(),
		watchedDirs: make(map[string]bool),
	}
	m.expanded[absPath] = true
	return m
}

func main() {
	m := initialModel()
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}