# CLAUDE.md

This file provides guidance for Claude Code when working with this project.

## Project Overview

Tiny Code Viewer is a terminal-based file browser and code viewer written in Go. It uses the Bubble Tea framework (Elm Architecture) for its TUI and provides a split-pane interface for navigating directory trees and previewing files with syntax highlighting.

## Architecture

### Design Pattern: Model-View-Update (MVU)

The application follows Bubble Tea's Elm Architecture:

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Model     │────▶│    Update   │────▶│    View     │
│  (State)    │     │  (Messages) │     │  (Render)   │
└─────────────┘     └─────────────┘     └─────────────┘
       ▲                                        │
       └────────────────────────────────────────┘
```

### Project Structure (Multi-file Architecture)

```
tiny-code-viewer/
├── main.go     # Application entry point and initialization
├── model.go    # Data structures and styling definitions
├── tree.go     # File tree building and flattening
├── file.go     # File loading and language detection
├── watcher.go  # File system watching and auto-refresh
├── update.go   # Message handling and state updates
├── view.go     # UI rendering
├── go.mod      # Go module definition
├── go.sum      # Dependency checksums
└── README.md   # Project documentation
```

### Core Data Structures

**fileNode** (`model.go:11-16`): Tree structure for file system representation
```go
type fileNode struct {
    name     string      // File/directory name
    path     string      // Absolute path
    isDir    bool        // Directory flag
    children []fileNode  // Child nodes (for directories)
}
```

**itemInfo** (`model.go:18-22`): Flattened item representation
```go
type itemInfo struct {
    path  string  // File/directory path
    isDir bool    // Directory flag
    depth int     // Nesting depth for indentation
}
```

**model** (`model.go:30-50`): Application state container
```go
type model struct {
    root          fileNode        // File tree root
    rootPath      string          // Starting directory path
    content       string          // Current file content
    filePath      string          // Current file path
    width, height int             // Terminal dimensions
    cursor        int             // Selection cursor index
    treeScroll    int             // Tree viewport scroll position
    previewScroll int             // Preview viewport scroll position
    focusLeft     bool            // Which panel has focus (true = tree)
    expanded      map[string]bool // Directory expansion state
    lastKeyTime   time.Time       // Key repeat rate limiting

    // File watching
    watcher       *fsnotify.Watcher  // File system watcher
    watchedDirs   map[string]bool    // Currently watched directories
    watchMu       sync.Mutex         // Watch operations mutex
    debounceMu    sync.Mutex         // Debounce operations mutex
    debounceTimer *time.Timer        // Debounce timer
}
```

**Messages** (`model.go:24-29`):
```go
type fsChangeEvent struct{}       // File system change event
type debouncedRefreshMsg struct{} // Debounced refresh message
```

## Key Functions

| Function | Location | Purpose |
|----------|----------|---------|
| `initialModel()` | `main.go:11-26` | Initializes application state |
| `buildTree()` | `tree.go:8-32` | Recursively builds file tree from root path |
| `flattenTree()` | `tree.go:34-45` | Flattens tree to linear list respecting expansion state |
| `getVisibleItems()` | `tree.go:47-49` | Returns flattened, visible tree items |
| `isBinaryFile()` | `file.go:10-29` | Detects binary files by checking for null bytes |
| `getLanguage()` | `file.go:31-74` | Maps file extensions to Chroma language identifiers |
| `loadFile()` | `file.go:76-94` | Loads and stores file content |
| `Update()` | `update.go:10-124` | Handles user input and state updates |
| `renderTree()` | `view.go:13-73` | Renders left panel (file tree) |
| `renderPreview()` | `view.go:75-123` | Renders right panel (syntax-highlighted preview) |
| `View()` | `view.go:125-202` | Composes final UI layout |
| `Init()` | `watcher.go:12-18` | Initializes file watcher |
| `waitForFsEvent()` | `watcher.go:21-46` | Waits for file system events |
| `watchDir()` | `watcher.go:49-69` | Adds directory to watcher |
| `updateWatches()` | `watcher.go:72-95` | Updates watched directories |
| `refreshTree()` | `watcher.go:123-162` | Rebuilds tree preserving state |

## Styling System

Uses Lipgloss for terminal styling (`model.go:53-74`):

- **dirStyle**: Light blue foreground (#39), bold - for directories
- **fileStyle**: Light gray foreground (#252) - for files
- **selectedStyle**: Light blue background (#63), white foreground (#0), bold - for selected items
- **titleStyle**: Blue foreground (#26), bold - for panel titles
- **connectorStyle**: Dark gray foreground (#240) - for tree branch connectors

### Tree Connectors

The file tree uses Unicode box-drawing characters (`├─`, `└─`, `│`) to display
a structured tree with visual branch lines. The root node (depth 0) is skipped
in the display since the path header already shows the current directory.

## Key Implementation Details

### 1. Lazy Loading
Files are loaded only when selected, not pre-loaded. This improves startup performance for large directories.

### 2. Viewport Scrolling
Both tree and preview panels use viewport-based rendering:
- Only visible lines are rendered
- `treeScroll` and `previewScroll` track scroll positions
- Content is sliced before rendering

### 3. Key Rate Limiting
Preview panel scrolling is rate-limited to 50ms (`update.go:34-44`) to prevent UI lag from rapid key presses.

### 4. Color Bleeding Prevention
ANSI reset codes (`\x1b[0m`) are appended to each line in the preview panel (`view.go:107-111`) to prevent syntax highlighting colors from bleeding into subsequent lines.

### 5. Binary File Detection
Checks first 512 bytes for null bytes (`file.go:17-28`). Binary files display `[Binary file - cannot preview]`.

### 6. Auto-refresh with Debouncing
File system changes are detected via fsnotify and debounced with 200ms delay (`watcher.go:98-120`) to avoid excessive refreshes during bulk operations.

## Development Guidelines

### Adding New Language Support

Edit `getLanguage()` function (`file.go:33-60`):

```go
langMap := map[string]string{
    // Add new extension mapping
    ".rs": "rust",
}
```

See [Chroma lexers](https://github.com/alecthomas/chroma/tree/master/lexers) for available languages.

### Modifying Key Bindings

Key handling is in `Update()` method (`update.go:46-117`). Key strings match `tea.KeyMsg.String()` output:
- Arrow keys: `"up"`, `"down"`, `"left"`, `"right"`
- Letters: `"k"`, `"j"`, etc.
- Special: `"enter"`, `"space"`, `"tab"`, `"ctrl+c"`

### Adjusting Panel Layout

Panel width calculation in `View()` (`view.go:130-136`):
```go
treeWidth := m.width / 3  // Tree takes 1/3 of width
if treeWidth < 25 { treeWidth = 25 }   // Minimum width
if treeWidth > 55 { treeWidth = 55 }   // Maximum width
```

### Changing Color Scheme

Modify style variables in `model.go` (`model.go:52-68`). Colors use ANSI 256-color codes. Use [this reference](https://www.ditig.com/256-colors-cheat-sheet) for color codes.

## Common Development Tasks

### Run Development Build
```bash
go run . [path]
```

### Build Release Binary
```bash
go build -ldflags="-s -w" -o tcv
```

### Run with Specific Directory
```bash
go run . ./src
```

## Dependencies

- [bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework (Elm Architecture)
- [lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [chroma](https://github.com/alecthomas/chroma) - Syntax highlighting engine
- [fsnotify](https://github.com/fsnotify/fsnotify) - File system notification

## Known Limitations

1. **No search functionality**: Files must be navigated manually
2. **No file editing**: View-only mode
3. **No hidden file toggle**: Hidden files are always visible
4. **No configuration file**: All settings are hardcoded