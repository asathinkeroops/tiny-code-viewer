package main

import (
	"time"

	"github.com/charmbracelet/bubbletea"
)

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case fsChangeEvent:
		// File system changed, schedule debounced refresh
		m.debounceMu.Lock()
		if m.debounceTimer != nil {
			m.debounceTimer.Stop()
		}
		m.debounceTimer = time.NewTimer(200 * time.Millisecond)
		m.debounceMu.Unlock()

		// Return command that waits for the timer
		return m, func() tea.Msg {
			time.Sleep(200 * time.Millisecond)
			return debouncedRefreshMsg{}
		}

	case debouncedRefreshMsg:
		// Actually perform the refresh
		m.refreshTree()
		// Continue listening for file events
		return m, m.waitForFsEvent()

	case tea.KeyMsg:
		// Limit key repeat rate for scroll operations
		now := time.Now()
		keyStr := msg.String()
		isScrollKey := keyStr == "up" || keyStr == "down" || keyStr == "k" || keyStr == "j"
		if isScrollKey && !m.focusLeft {
			// Only apply delay limit for preview scrolling
			if now.Sub(m.lastKeyTime) < 50*time.Millisecond {
				return m, nil
			}
		}
		m.lastKeyTime = now

		switch keyStr {
		case "ctrl+c", "q":
			// Clean up watcher before quitting
			if m.watcher != nil {
				m.watcher.Close()
			}
			return m, tea.Quit
		case "tab":
			m.focusLeft = !m.focusLeft
		case "r":
			// Manual refresh
			m.refreshTree()
		case "up", "k":
			if m.focusLeft {
				if m.cursor > 1 {
					m.cursor--
					if m.cursor < m.treeScroll {
						m.treeScroll = m.cursor
					}
				}
			} else {
				if m.previewScroll > 0 {
					m.previewScroll--
				}
			}
		case "down", "j":
			if m.focusLeft {
				items := m.getVisibleItems()
				if m.cursor < len(items)-1 {
					m.cursor++
					contentHeight := m.panelHeight() - 1
					if m.cursor >= m.treeScroll+contentHeight {
						m.treeScroll = m.cursor - contentHeight + 1
					}
				}
			} else {
				m.previewScroll++
			}
		case "enter", " ":
			items := m.getVisibleItems()
			if m.cursor < len(items) {
				item := items[m.cursor]
				if item.isDir {
					// Toggle expansion
					if !m.expanded[item.path] {
						// Expanding: load children if not loaded
						m.loadDirChildren(item.path)
					}
					m.expanded[item.path] = !m.expanded[item.path]
				} else {
					m.loadFile(item.path)
					m.focusLeft = false // Switch to preview after opening file
				}
			}
		case "left", "h":
			items := m.getVisibleItems()
			if m.cursor < len(items) {
				item := items[m.cursor]
				if item.isDir && m.expanded[item.path] {
					m.expanded[item.path] = false
				}
			}
		case "right", "l":
			items := m.getVisibleItems()
			if m.cursor < len(items) {
				item := items[m.cursor]
				if item.isDir {
					// Load children before expanding
					m.loadDirChildren(item.path)
					m.expanded[item.path] = true
				} else {
					m.loadFile(item.path)
				}
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// loadDirChildren finds a directory node by path and loads its children.
// This is used for lazy loading when a directory is expanded.
func (m *model) loadDirChildren(dirPath string) {
	node := findNode(&m.root, dirPath)
	if node != nil {
		node.loadChildren()
	}
}

// findNode recursively finds a node by path in the tree.
func findNode(node *fileNode, path string) *fileNode {
	if node.path == path {
		return node
	}
	for i := range node.children {
		if found := findNode(&node.children[i], path); found != nil {
			return found
		}
	}
	return nil
}
