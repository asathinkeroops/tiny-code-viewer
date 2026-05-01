package tcv

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/fsnotify/fsnotify"
)

type fileNode struct {
	name     string
	path     string
	isDir    bool
	loaded   bool
	children []fileNode
}

type itemInfo struct {
	path  string
	isDir bool
	depth int
}

type fsChangeEvent struct{}
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

	highlightedLines []string
	highlightedPath  string

	watcher       *fsnotify.Watcher
	watchedDirs   map[string]bool
	watchMu       sync.Mutex
	debounceMu    sync.Mutex
	debounceTimer *time.Timer
}

func NewModel() *model {
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
		cursor:      1,
		watcher:     initialWatcher(),
		watchedDirs: make(map[string]bool),
	}
	m.expanded[absPath] = true
	return m
}

var (
	dirStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true)

	fileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("51")).
			Bold(true)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("26"))

	connectorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
)
