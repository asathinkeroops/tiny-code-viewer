package main

import (
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/fsnotify/fsnotify"
)

type fileNode struct {
	name     string
	path     string
	isDir    bool
	children []fileNode
}

type itemInfo struct {
	path  string
	isDir bool
	depth int
}

// File system change message with debouncing
type fsChangeEvent struct{}

// Debounced refresh message
type debouncedRefreshMsg struct{}

type model struct {
	root          fileNode
	rootPath      string
	content       string
	filePath      string
	width         int
	height        int
	cursor        int
	treeScroll    int
	previewScroll int
	focusLeft     bool
	expanded      map[string]bool
	lastKeyTime   time.Time

	// File watching
	watcher       *fsnotify.Watcher
	watchedDirs   map[string]bool
	watchMu       sync.Mutex
	debounceMu    sync.Mutex
	debounceTimer *time.Timer
}

var (
	dirStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("34")).
		Bold(true)

	fileStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("244"))

	selectedStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("63")).
		Foreground(lipgloss.Color("0")).
		Bold(true)

	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("26"))
)