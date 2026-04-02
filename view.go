package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/charmbracelet/lipgloss"
)

func (m model) renderTree() string {
	items := m.getVisibleItems()

	// Path header - replace HOME with ~
	displayPath := m.rootPath
	homeDir, err := os.UserHomeDir()
	if err == nil && strings.HasPrefix(displayPath, homeDir) {
		displayPath = "~" + strings.TrimPrefix(displayPath, homeDir)
	}
	pathStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("136"))
	header := pathStyle.Render(displayPath)

	// Use same height calculation as View
	panelHeight := m.height - 3
	if panelHeight < 1 {
		panelHeight = 10
	}
	// Account for header line
	contentHeight := panelHeight - 1

	var lines []string
	end := min(m.treeScroll+contentHeight, len(items))

	for i := m.treeScroll; i < end; i++ {
		item := items[i]
		name := filepath.Base(item.path)

		// Add indentation based on depth
		indent := strings.Repeat("  ", item.depth)

		var prefix string
		if item.isDir {
			if m.expanded[item.path] {
				prefix = "▼ "
			} else {
				prefix = "▶ "
			}
		} else {
			prefix = ""
		}

		line := indent + prefix + name

		if i == m.cursor && m.focusLeft {
			lines = append(lines, selectedStyle.Render(line))
		} else if i == m.cursor {
			// Cursor visible but not focused
			style := lipgloss.NewStyle().Background(lipgloss.Color("238"))
			if item.isDir {
				lines = append(lines, style.Foreground(lipgloss.Color("34")).Bold(true).Render(line))
			} else {
				lines = append(lines, style.Foreground(lipgloss.Color("244")).Render(line))
			}
		} else {
			if item.isDir {
				lines = append(lines, dirStyle.Render(line))
			} else {
				lines = append(lines, fileStyle.Render(line))
			}
		}
	}

	return header + "\n" + strings.Join(lines, "\n")
}

func (m model) renderPreview() string {
	if m.content == "" {
		return "  Select a file to preview"
	}

	// Highlight code
	var buf bytes.Buffer
	lang := getLanguage(m.filePath)
	err := quick.Highlight(&buf, m.content, lang, "terminal256", "friendly")
	if err != nil {
		buf.WriteString(m.content)
	}

	highlighted := buf.String()
	lines := strings.Split(highlighted, "\n")

	// Use same height calculation as View
	panelHeight := m.height - 3
	if panelHeight < 1 {
		panelHeight = 10
	}
	// Account for title line
	contentHeight := panelHeight - 1

	// Slice content for scroll
	start := m.previewScroll
	end := min(start+contentHeight, len(lines))
	if start > end {
		start = 0
	}

	// Add reset code at end of each line to prevent color bleeding
	resetCode := "\x1b[0m"
	visibleLines := lines[start:end]
	for i, line := range visibleLines {
		visibleLines[i] = line + resetCode
	}

	title := titleStyle.Render(filepath.Base(m.filePath))
	content := strings.Join(visibleLines, "\n")

	// Add scroll indicator if content is longer than panel
	if len(lines) > contentHeight {
		scrollInfo := fmt.Sprintf(" [%d/%d lines]", start+1, len(lines))
		title = title + scrollInfo
	}

	return title + "\n" + content + resetCode
}

func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	treeWidth := m.width / 3
	if treeWidth < 25 {
		treeWidth = 25
	}
	if treeWidth > 55 {
		treeWidth = 55
	}
	panelHeight := m.height - 3

	// Left panel - truncate content to panelHeight lines
	leftContent := m.renderTree()
	leftLines := strings.Split(leftContent, "\n")
	if len(leftLines) > panelHeight {
		leftLines = leftLines[:panelHeight]
	}
	leftTruncated := strings.Join(leftLines, "\n")

	// Right panel - truncate content to panelHeight lines
	rightContent := m.renderPreview()
	rightLines := strings.Split(rightContent, "\n")
	if len(rightLines) > panelHeight {
		rightLines = rightLines[:panelHeight]
	}
	rightTruncated := strings.Join(rightLines, "\n")

	// Build panels - both Height and MaxHeight to fill and cap at panelHeight
	leftPanel := lipgloss.NewStyle().
		Width(treeWidth).
		Height(panelHeight).
		MaxHeight(panelHeight).
		BorderRight(true).
		BorderStyle(lipgloss.NormalBorder()).
		PaddingLeft(1).
		Render(leftTruncated)

	rightPanel := lipgloss.NewStyle().
		Width(m.width - treeWidth - 2).
		Height(panelHeight).
		MaxHeight(panelHeight).
		PaddingLeft(1).
		Render(rightTruncated)

	panels := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

	// Help bar
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	sepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	focusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true)

	focusLabel := "Tree"
	if !m.focusLeft {
		focusLabel = "Preview"
	}
	focusTag := " [" + focusStyle.Render(focusLabel) + "]"

	// Build styled help items
	helpItems := []string{
		keyStyle.Render("↑/k") + descStyle.Render(" Up"),
		keyStyle.Render("↓/j") + descStyle.Render(" Down"),
		keyStyle.Render("←/h") + descStyle.Render(" Collapse"),
		keyStyle.Render("→/l") + descStyle.Render(" Expand"),
		keyStyle.Render("r") + descStyle.Render(" Refresh"),
		keyStyle.Render("Tab") + descStyle.Render(" Switch") + focusTag,
		keyStyle.Render("Enter") + descStyle.Render(" Open"),
		keyStyle.Render("q") + descStyle.Render(" Quit"),
	}
	helpLine := sepStyle.Render(" │ ") + strings.Join(helpItems, sepStyle.Render(" │ "))

	helpBar := helpLine

	return panels + "\n" + helpBar
}