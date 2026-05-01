package tcv

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
)

func (m *model) Init() tea.Cmd {
	if m.watcher != nil {
		m.watchDir(m.rootPath)
	}
	return m.waitForFsEvent()
}

func (m *model) waitForFsEvent() tea.Cmd {
	if m.watcher == nil {
		return nil
	}
	return func() tea.Msg {
		select {
		case event, ok := <-m.watcher.Events:
			if !ok {
				return nil
			}
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

func (m *model) watchDir(path string) {
	if m.watcher == nil {
		return
	}

	m.watchMu.Lock()
	defer m.watchMu.Unlock()

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

func (m *model) updateWatches() {
	if m.watcher == nil {
		return
	}

	m.watchMu.Lock()
	defer m.watchMu.Unlock()

	if !m.watchedDirs[m.rootPath] {
		if err := m.watcher.Add(m.rootPath); err == nil {
			m.watchedDirs[m.rootPath] = true
		}
	}

	for path, expanded := range m.expanded {
		if expanded && !m.watchedDirs[path] {
			if err := m.watcher.Add(path); err == nil {
				m.watchedDirs[path] = true
			}
		}
	}
}

func (m *model) triggerDebouncedRefresh() tea.Cmd {
	return func() tea.Msg {
		m.debounceMu.Lock()
		defer m.debounceMu.Unlock()

		if m.debounceTimer != nil {
			m.debounceTimer.Stop()
		}

		timer := time.NewTimer(200 * time.Millisecond)
		m.debounceTimer = timer

		go func() {
			<-timer.C
		}()

		return debouncedRefreshMsg{}
	}
}

func (m *model) refreshTree() {
	items := m.getVisibleItems()
	var selectedPath string
	if m.cursor < len(items) {
		selectedPath = items[m.cursor].path
	}

	expandedPaths := make(map[string]bool)
	for path, expanded := range m.expanded {
		if expanded {
			expandedPaths[path] = true
		}
	}

	m.root = buildTree(m.rootPath)

	for path := range expandedPaths {
		node := findNode(&m.root, path)
		if node != nil && node.isDir {
			node.loadChildren()
		}
	}

	newItems := m.getVisibleItems()
	for i, item := range newItems {
		if item.path == selectedPath {
			m.cursor = i
			contentHeight := m.height - 3
			if m.cursor < m.treeScroll {
				m.treeScroll = m.cursor
			} else if m.cursor >= m.treeScroll+contentHeight {
				m.treeScroll = m.cursor - contentHeight + 1
			}
			break
		}
	}

	m.updateWatches()

	if m.filePath != "" {
		if _, err := os.Stat(m.filePath); err == nil {
			m.loadFile(m.filePath)
		} else {
			m.content = "[File has been deleted]"
		}
	}
}

func initialWatcher() *fsnotify.Watcher {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("Warning: could not create file watcher: %v\n", err)
		return nil
	}
	return watcher
}
