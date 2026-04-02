package main

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
)

func (m model) Init() tea.Cmd {
	// Start watching the root directory
	if m.watcher != nil {
		m.watchDir(m.rootPath)
	}
	return m.waitForFsEvent()
}

// waitForFsEvent returns a command that waits for file system events
func (m model) waitForFsEvent() tea.Cmd {
	if m.watcher == nil {
		return nil
	}
	return func() tea.Msg {
		select {
		case event, ok := <-m.watcher.Events:
			if !ok {
				return nil
			}
			// Filter out frequent events like chmod
			if event.Op&fsnotify.Create != 0 ||
				event.Op&fsnotify.Remove != 0 ||
				event.Op&fsnotify.Rename != 0 ||
				event.Op&fsnotify.Write != 0 {
				return fsChangeEvent{}
			}
			return nil
		case _, ok := <-m.watcher.Errors:
			if !ok {
				return nil
			}
			return nil
		}
	}
}

// watchDir adds a directory to the watcher (thread-safe)
func (m *model) watchDir(path string) {
	if m.watcher == nil {
		return
	}

	m.watchMu.Lock()
	defer m.watchMu.Unlock()

	// Only watch expanded directories
	if !m.expanded[path] && path != m.rootPath {
		return
	}

	if m.watchedDirs[path] {
		return
	}

	if err := m.watcher.Add(path); err == nil {
		m.watchedDirs[path] = true
	}
}

// updateWatches ensures all expanded directories are watched
func (m *model) updateWatches() {
	if m.watcher == nil {
		return
	}

	m.watchMu.Lock()
	defer m.watchMu.Unlock()

	// Watch root
	if !m.watchedDirs[m.rootPath] {
		if err := m.watcher.Add(m.rootPath); err == nil {
			m.watchedDirs[m.rootPath] = true
		}
	}

	// Watch expanded directories
	for path, expanded := range m.expanded {
		if expanded && !m.watchedDirs[path] {
			if err := m.watcher.Add(path); err == nil {
				m.watchedDirs[path] = true
			}
		}
	}
}

// triggerDebouncedRefresh refreshes the tree with debouncing
func (m *model) triggerDebouncedRefresh() tea.Cmd {
	return func() tea.Msg {
		m.debounceMu.Lock()
		defer m.debounceMu.Unlock()

		// Cancel existing timer if any
		if m.debounceTimer != nil {
			m.debounceTimer.Stop()
		}

		// Wait 200ms before actually refreshing
		timer := time.NewTimer(200 * time.Millisecond)
		m.debounceTimer = timer

		go func() {
			<-timer.C
			// Send refresh message through the program
			// We'll handle this in Update
		}()

		return debouncedRefreshMsg{}
	}
}

// refreshTree rebuilds the tree while preserving state
func (m *model) refreshTree() {
	// Remember current selection
	items := m.getVisibleItems()
	var selectedPath string
	if m.cursor < len(items) {
		selectedPath = items[m.cursor].path
	}

	// Rebuild tree
	m.root = buildTree(m.rootPath)

	// Restore cursor position
	newItems := m.getVisibleItems()
	for i, item := range newItems {
		if item.path == selectedPath {
			m.cursor = i
			// Adjust scroll if needed
			contentHeight := m.height - 4
			if m.cursor < m.treeScroll {
				m.treeScroll = m.cursor
			} else if m.cursor >= m.treeScroll+contentHeight {
				m.treeScroll = m.cursor - contentHeight + 1
			}
			break
		}
	}

	// Update watches for new directories
	m.updateWatches()

	// Reload file if currently viewing one
	if m.filePath != "" {
		if _, err := os.Stat(m.filePath); err == nil {
			m.loadFile(m.filePath)
		} else {
			// File was deleted
			m.content = "[File has been deleted]"
		}
	}
}

// initialWatcher creates a new file watcher
func initialWatcher() *fsnotify.Watcher {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("Warning: could not create file watcher: %v\n", err)
		return nil
	}
	return watcher
}