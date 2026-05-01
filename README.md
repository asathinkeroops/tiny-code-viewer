# Tiny Code Viewer

![Screenshot](snapshots/screen.png)

A lightweight terminal-based file browser and code viewer with syntax highlighting, built with Go and Bubble Tea.

This tool is designed to work alongside [Claude Code](https://claude.ai/code), providing a convenient way to browse project directory structure and view source code directly in the terminal. It complements Claude Code's capabilities by offering a visual file explorer for quick code navigation.

## Features

- **Split-pane Interface**: Browse directory trees on the left, preview files on the right
- **Syntax Highlighting**: Support for 23 programming languages using Chroma
- **Line Numbers**: Line numbers in the preview panel for easy reference
- **Vim-style Navigation**: Intuitive keyboard shortcuts (h/j/k/l, g/G, arrow keys)
- **Binary File Detection**: Automatically detects and skips binary files
- **Lazy Loading**: Directory children loaded on demand for fast startup
- **Mouse Wheel Support**: Scroll the preview pane with the mouse wheel
- **Scrollbar**: Visual scroll bar in the preview pane
- **Responsive Layout**: Adapts to terminal window size
- **Auto Refresh**: Automatically updates file tree when files are added, removed, or renamed
- **Debounced Updates**: 200ms delay to prevent excessive refreshes during bulk operations
- **Fast and Lightweight**: Single binary with minimal external dependencies

## Supported Languages

Go, Python, JavaScript, TypeScript, JSX, TSX, Java, C, C++, Rust, Ruby, PHP, Bash, JSON, YAML, TOML, XML, HTML, CSS, SQL, Markdown, Dockerfile, Makefile

## Installation

### Go Install

```bash
go install github.com/asathinkeroops/tiny-code-viewer/cmd/tcv@latest

## run
tcv
```

### Build from Source

```bash
# Clone the repository
git clone https://github.com/asathinkeroops/tiny-code-viewer.git
cd tiny-code-viewer

# Build
go build -o tcv ./cmd/tcv

# Optional: Install to $GOPATH/bin
go install ./cmd/tcv
```

### Requirements

- Go 1.24.0 or later

## Usage

```bash
# View current directory
./tcv

# View a specific directory
./tcv /path/to/project

# View a specific file's parent directory
./tcv /path/to/file.go
```

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `↑` / `k` | Move cursor up |
| `↓` / `j` | Move cursor down |
| `←` / `h` | Collapse directory |
| `→` / `l` | Expand directory / Open file |
| `Enter` / `Space` | Toggle directory / Open file |
| `r` | Refresh file tree (manual) |
| `Tab` | Switch focus between tree and preview panels |
| `q` / `Ctrl+C` | Quit |

**Preview panel shortcuts** (when focused):

| Key | Action |
|-----|--------|
| `↑` / `k` | Scroll up one line |
| `↓` / `j` | Scroll down one line |
| `PgUp` / `Ctrl+U` | Scroll up half page |
| `PgDn` / `Ctrl+D` | Scroll down half page |
| `Home` / `g` | Jump to top |
| `End` / `G` | Jump to bottom |
| Mouse wheel | Scroll up / down 3 lines |

## Auto Refresh

The viewer automatically watches for file system changes:
- **Auto-refresh**: File tree updates automatically when files are added, removed, or renamed
- **Debouncing**: Changes are batched with 200ms delay to avoid excessive refreshes during bulk operations
- **Smart watching**: Only expanded directories are watched for better performance

## User Interface

```
┌──────────────────────────────────────────────────────────────────────┐
│ ~/project                    │  main.go [1/150 lines]                │
│ ▼ src                        │  1 │ package main                     │
│   ▶ cmd                      │  2 │                                  │
│   ▼ internal                 │  3 │ func main() {                    │
│     ├─ main.go               │  4 │     m := tcv.NewModel()          │
│     └─ config.go             │  5 │     p := tea.NewProgram(m)       │
│ ▼ pkg                        │  6 │     // ...                       │
│   └─ utils.go                │  7 │ }                                │
│                              │    │                                  │
├──────────────────────────────────────────────────────────────────────┤
│ ↑/k Up │ ↓/j Down │ ←/h Collapse │ →/l Expand │ ... │ Tab Switch [Tree]
└──────────────────────────────────────────────────────────────────────┘
```

## Dependencies

- [bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework (Elm Architecture)
- [lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [chroma](https://github.com/alecthomas/chroma) - Syntax highlighting engine
- [fsnotify](https://github.com/fsnotify/fsnotify) - File system notification
- [go-runewidth](https://github.com/mattn/go-runewidth) - Handles CJK and wide character widths

## Project Structure

```
tiny-code-viewer/
├── cmd/
│   └── tcv/
│       └── main.go        # Application entry point
├── internal/
│   └── tcv/
│       ├── model.go       # Model, messages, and lipgloss styles
│       ├── tree.go        # File tree with lazy loading
│       ├── file.go        # File loading, binary detection, syntax highlighting
│       ├── watcher.go     # fsnotify watcher, debounced auto-refresh
│       ├── update.go      # Key/mouse/window event handling (Bubble Tea Update)
│       └── view.go        # Split-pane UI, line numbers, scrollbar rendering
├── go.mod                 # Go module definition
├── go.sum                 # Dependency checksums
└── README.md              # This file
```

## License

MIT License