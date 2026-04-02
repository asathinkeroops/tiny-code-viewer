package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func isBinaryFile(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return true
	}
	defer file.Close()

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil {
		return true
	}

	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return true
		}
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
	m.previewScroll = 0 // Reset scroll when loading new file

	if isBinaryFile(path) {
		m.content = "[Binary file - cannot preview]"
		m.filePath = path
		return
	}

	data, err := os.ReadFile(path)
	if err != nil {
		m.content = fmt.Sprintf("Error reading file: %v", err)
		m.filePath = path
		return
	}

	m.content = string(data)
	m.filePath = path
}