package main

import (
	"time"

	"github.com/charmbracelet/bubbletea"
)

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case fsChangeEvent:
		m.debounceMu.Lock()
		if m.debounceTimer != nil {
			m.debounceTimer.Stop()
		}
		m.debounceTimer = time.NewTimer(200 * time.Millisecond)
		m.debounceMu.Unlock()

		return m, func() tea.Msg {
			time.Sleep(200 * time.Millisecond)
			return debouncedRefreshMsg{}
		}

	case debouncedRefreshMsg:
		m.refreshTree()
		return m, m.waitForFsEvent()

	case tea.MouseMsg:
		if !m.focusLeft && m.content != "" {
			scrollAmount := 3
			if scrollAmount < 1 {
				scrollAmount = 1
			}
			switch msg.Button {
			case tea.MouseButtonWheelUp:
				m.previewScroll -= scrollAmount
				if m.previewScroll < 0 {
					m.previewScroll = 0
				}
			case tea.MouseButtonWheelDown:
				m.previewScroll += scrollAmount
			}
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.watcher != nil {
				m.watcher.Close()
			}
			return m, tea.Quit
		case "tab":
			m.focusLeft = !m.focusLeft
		case "r":
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
		case "pgup", "ctrl+u":
			if !m.focusLeft && m.content != "" {
				pageSize := m.panelHeight() / 2
				if pageSize < 1 {
					pageSize = 1
				}
				m.previewScroll -= pageSize
				if m.previewScroll < 0 {
					m.previewScroll = 0
				}
			}
		case "pgdown", "ctrl+d":
			if !m.focusLeft && m.content != "" {
				pageSize := m.panelHeight() / 2
				if pageSize < 1 {
					pageSize = 1
				}
				m.previewScroll += pageSize
			}
		case "home", "g":
			if !m.focusLeft && m.content != "" {
				m.previewScroll = 0
			}
		case "end", "G":
			if !m.focusLeft && m.content != "" {
				m.previewScroll = 1 << 30
			}
		case "enter", " ":
			items := m.getVisibleItems()
			if m.cursor < len(items) {
				item := items[m.cursor]
				if item.isDir {
					if !m.expanded[item.path] {
						m.loadDirChildren(item.path)
					}
					m.expanded[item.path] = !m.expanded[item.path]
				} else {
					m.loadFile(item.path)
					m.focusLeft = false
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
