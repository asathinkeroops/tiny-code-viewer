package tcv

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/v2/quick"
)

func isBinaryFile(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return true
	}
	defer file.Close()

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil || n == 0 {
		return true
	}
	buf = buf[:n]

	if n >= 2 {
		if (buf[0] == 0xFF && buf[1] == 0xFE) || (buf[0] == 0xFE && buf[1] == 0xFF) {
			return false
		}
	}
	if n >= 3 && buf[0] == 0xEF && buf[1] == 0xBB && buf[2] == 0xBF {
		return false
	}

	nullCount := 0
	nonNullPrintable := 0
	nonNullTotal := 0
	for _, b := range buf {
		if b == 0 {
			nullCount++
		} else {
			nonNullTotal++
			if (b >= 0x20 && b <= 0x7E) || b == 0x09 || b == 0x0A || b == 0x0D {
				nonNullPrintable++
			}
		}
	}

	if nullCount > 0 && float64(nullCount)/float64(n) >= 0.4 {
		if nonNullTotal > 0 && float64(nonNullPrintable)/float64(nonNullTotal) >= 0.8 {
			return false
		}
		return true
	}

	if nullCount > 0 {
		return true
	}

	return false
}

func getLanguage(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	langMap := map[string]string{
		".go":   "go",
		".py":   "python",
		".js":   "javascript",
		".ts":   "typescript",
		".jsx":  "jsx",
		".tsx":  "tsx",
		".java": "java",
		".c":    "c",
		".cpp":  "cpp",
		".h":    "c",
		".hpp":  "cpp",
		".rs":   "rust",
		".rb":   "ruby",
		".php":  "php",
		".sh":   "bash",
		".bash": "bash",
		".zsh":  "bash",
		".json": "json",
		".yaml": "yaml",
		".yml":  "yaml",
		".toml": "toml",
		".xml":  "xml",
		".html": "html",
		".css":  "css",
		".sql":  "sql",
		".md":   "markdown",
	}

	base := strings.ToLower(filepath.Base(filename))
	if base == "dockerfile" || strings.HasPrefix(base, "dockerfile.") {
		return "dockerfile"
	}
	if base == "makefile" {
		return "makefile"
	}

	if lang, ok := langMap[ext]; ok {
		return lang
	}
	return "text"
}

func (m *model) loadFile(path string) {
	m.previewScroll = 0
	m.highlightedLines = nil
	m.highlightedPath = ""

	if isBinaryFile(path) {
		m.content = "[Binary file - cannot preview]"
		m.filePath = path
		m.highlightedLines = strings.Split(m.content, "\n")
		m.highlightedPath = path
		return
	}

	data, err := os.ReadFile(path)
	if err != nil {
		m.content = fmt.Sprintf("Error reading file: %v", err)
		m.filePath = path
		m.highlightedLines = strings.Split(m.content, "\n")
		m.highlightedPath = path
		return
	}

	m.content = string(data)
	m.filePath = path

	lang := getLanguage(path)
	var buf bytes.Buffer
	err = quick.Highlight(&buf, m.content, lang, "terminal256", "friendly")
	if err != nil {
		m.highlightedLines = strings.Split(m.content, "\n")
	} else {
		highlighted := buf.String()
		m.highlightedLines = strings.Split(highlighted, "\n")
		if len(m.highlightedLines) > 0 && m.highlightedLines[len(m.highlightedLines)-1] == "" {
			m.highlightedLines = m.highlightedLines[:len(m.highlightedLines)-1]
		}
	}
	m.highlightedPath = path
}
