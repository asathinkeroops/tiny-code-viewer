package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

func (m *model) renderTree() string {
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

	panelHeight := m.panelHeight()
	contentHeight := panelHeight - 1

	treeWidth := m.treePanelWidth()
	maxLineWidth := treeWidth - 2

	var lines []string
	end := min(m.treeScroll+contentHeight, len(items))

	for i := m.treeScroll; i < end; i++ {
		item := items[i]

		// Skip root node (depth 0) - header already shows the path
		if item.depth == 0 {
			continue
		}

		name := filepath.Base(item.path)

		// Build tree connector prefix for this item
		connector := treeConnectorPrefix(items, i, item.depth)

		var expandIcon string
		if item.isDir {
			if m.expanded[item.path] {
				expandIcon = "▼ "
			} else {
				expandIcon = "▶ "
			}
		}

		// Truncate name to fit within panel width
		prefixLen := len(connector + expandIcon)
		available := maxLineWidth - prefixLen
		if available < 3 {
			available = 3
		}
		name = truncateStr(name, available)

		prefix := connectorStyle.Render(connector + expandIcon)

		if i == m.cursor && m.focusLeft {
			lines = append(lines, prefix+selectedStyle.Render(name))
		} else if i == m.cursor {
			style := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
			if item.isDir {
				lines = append(lines, prefix+style.Bold(true).Render(name))
			} else {
				lines = append(lines, prefix+style.Render(name))
			}
		} else {
			if item.isDir {
				lines = append(lines, dirStyle.Render(prefix+name))
			} else {
				lines = append(lines, fileStyle.Render(prefix+name))
			}
		}
	}

	return header + "\n" + strings.Join(lines, "\n")
}

func (m *model) panelHeight() int {
	helpBarWidth := m.helpBarVisualWidth()
	helpBarLines := 1
	if m.width > 0 && helpBarWidth > m.width {
		helpBarLines = (helpBarWidth + m.width - 1) / m.width
	}
	h := m.height - helpBarLines
	if h < 1 {
		h = 10
	}
	return h
}

func (m *model) helpBarVisualWidth() int {
	sep := " │ "
	items := []string{
		"↑/k Up", "↓/j Down", "←/h Collapse", "→/l Expand",
		"r Refresh", "Tab Switch [Preview]", "Enter Open", "q Quit",
	}
	return lipgloss.Width(sep + strings.Join(items, sep))
}

func (m *model) treePanelWidth() int {
	w := m.width / 3
	if w < 25 {
		w = 25
	}
	if w > 55 {
		w = 55
	}
	return w
}

func (m *model) previewTextWidth() int {
	treeW := m.treePanelWidth()
	// right panel width: m.width - treeW - 1 (subtract border only)
	// minus PaddingLeft(1) = actual text area
	w := m.width - treeW - 2
	if w < 10 {
		w = 10
	}
	return w
}

// truncateStr truncates a string to maxLen visual characters, adding "…" if truncated.
func truncateStr(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-1]) + "…"
}

// treeConnectorPrefix builds the connector string for an item based on its
// position in the tree. Uses box-drawing characters:
//
//	├─  for a non-last child, └─  for the last child
//	│   for an ancestor that still has more siblings below
func treeConnectorPrefix(items []itemInfo, idx int, depth int) string {
	var b strings.Builder

	// Ancestor connectors: for each level 1 to depth-1
	for level := 1; level < depth; level++ {
		if hasLaterSiblingAtLevel(items, idx, level) {
			b.WriteString("│  ")
		} else {
			b.WriteString("   ")
		}
	}

	// Item connector at current depth
	if hasLaterSiblingAtLevel(items, idx, depth) {
		b.WriteString("├─ ")
	} else {
		b.WriteString("└─ ")
	}

	return b.String()
}

// hasLaterSiblingAtLevel checks whether any item after idx has the given depth,
// stopping if a shallower depth is encountered first.
func hasLaterSiblingAtLevel(items []itemInfo, idx int, level int) bool {
	for j := idx + 1; j < len(items); j++ {
		if items[j].depth < level {
			return false
		}
		if items[j].depth == level {
			return true
		}
	}
	return false
}

func isANSITerminator(b byte) bool {
	return b >= 0x40 && b <= 0x7E
}

func extractPrefixSGR(line string) string {
	if len(line) < 2 || line[0] != '\x1b' {
		return ""
	}
	var b strings.Builder
	i := 0
	for i < len(line) && line[i] == '\x1b' {
		j := i + 1
		for j < len(line) && !isANSITerminator(line[j]) {
			j++
		}
		if j >= len(line) {
			break
		}
		j++
		if line[j-1] == 'm' {
			b.WriteString(line[i:j])
		}
		i = j
	}
	return b.String()
}

func wrapLineToWidth(line string, maxWidth int) []string {
	if maxWidth <= 0 {
		return []string{line}
	}
	if lipgloss.Width(line) <= maxWidth {
		return []string{line}
	}

	prefixSGR := extractPrefixSGR(line)

	var result []string
	var current strings.Builder
	col := 0
	i := 0

	for i < len(line) {
		if line[i] == '\x1b' {
			start := i
			j := i + 1
			for j < len(line) && !isANSITerminator(line[j]) {
				j++
			}
			if j < len(line) {
				j++
			}
			current.WriteString(line[start:j])
			i = j
			continue
		}

		r, size := utf8.DecodeRuneInString(line[i:])
		rw := runewidth.RuneWidth(r)

		if col+rw > maxWidth && col > 0 {
			result = append(result, current.String())
			current.Reset()
			if prefixSGR != "" {
				current.WriteString(prefixSGR)
			}
			col = 0
			continue
		}

		current.WriteRune(r)
		col += rw
		i += size
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}
	if len(result) == 0 {
		result = append(result, "")
	}

	return result
}

func (m *model) renderPreview() string {
	if m.content == "" {
		return "  Select a file to preview"
	}

	lines := m.highlightedLines
	if lines == nil {
		return "  Preview not available"
	}

	textWidth := m.previewTextWidth()

	gutterWidth := numWidth(len(lines)) + lineNumSepWidth
	codeWidth := textWidth - 1 - gutterWidth
	if codeWidth < 1 {
		codeWidth = 1
	}
	contentHeight := m.panelHeight() - 1
	if contentHeight < 1 {
		contentHeight = 1
	}

	type wrappedLine struct {
		text    string
		lineNum int
		isFirst bool
	}

	var wrappedLines []wrappedLine
	for lineNum, line := range lines {
		wrapped := wrapLineToWidth(line, codeWidth)
		for i, wl := range wrapped {
			wrappedLines = append(wrappedLines, wrappedLine{
				text:    wl,
				lineNum: lineNum + 1,
				isFirst: i == 0,
			})
		}
	}

	maxScroll := len(wrappedLines) - contentHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	start := m.previewScroll
	if start > maxScroll {
		start = maxScroll
		m.previewScroll = start
	}
	if start < 0 {
		start = 0
	}

	end := start + contentHeight
	if end > len(wrappedLines) {
		end = len(wrappedLines)
	}

	resetCode := "\x1b[0m"
	scrollBar := renderScrollBar(start, end, len(wrappedLines), contentHeight)

	visibleLines := make([]string, 0, end-start)
	contentWidth := textWidth - 1
	contentStyle := lipgloss.NewStyle().Width(contentWidth)
	for i := start; i < end; i++ {
		wl := wrappedLines[i]

		gutter := ""
		if wl.isFirst {
			gutter = fmtLineNum(wl.lineNum, len(lines))
		} else {
			gutter = lineNumBlank(len(lines))
		}

		codeLine := gutter + resetCode + wl.text + resetCode

		bar := ""
		if i-start < len(scrollBar) {
			bar = scrollBar[i-start]
		}
		visibleLines = append(visibleLines, contentStyle.Render(codeLine)+bar)
	}

	titleText := filepath.Base(m.filePath)
	if len(wrappedLines) > contentHeight {
		sourceStart := start + 1
		scrollInfo := fmt.Sprintf(" [%d/%d lines]", sourceStart, len(wrappedLines))
		titleText = titleText + scrollInfo
	}
	titleText = truncateStr(titleText, textWidth)
	title := titleStyle.Render(titleText)

	content := strings.Join(visibleLines, "\n")
	return title + "\n" + content + resetCode
}

func (m *model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	treeWidth := m.treePanelWidth()
	panelHeight := m.panelHeight()

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
		Width(m.width - treeWidth - 1).
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

var (
	scrollTrackStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("235"))
	scrollThumbStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

func renderScrollBar(start, end, total, height int) []string {
	if total <= height {
		return nil
	}

	bar := make([]string, height)

	thumbStart := int(float64(start) / float64(total) * float64(height))
	thumbEnd := int(float64(end) / float64(total) * float64(height))
	if thumbEnd <= thumbStart {
		thumbEnd = thumbStart + 1
	}
	if thumbStart < 0 {
		thumbStart = 0
	}
	if thumbEnd > height {
		thumbEnd = height
	}

	for i := 0; i < height; i++ {
		if i >= thumbStart && i < thumbEnd {
			bar[i] = scrollThumbStyle.Render("█")
		} else {
			bar[i] = scrollTrackStyle.Render("│")
		}
	}

	return bar
}

var (
	lineNumStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	lineNumSepStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
)

const lineNumSepWidth = 3

func numWidth(n int) int {
	if n < 10 {
		return 1
	}
	if n < 100 {
		return 2
	}
	if n < 1000 {
		return 3
	}
	if n < 10000 {
		return 4
	}
	if n < 100000 {
		return 5
	}
	return 6
}

func fmtLineNum(num, total int) string {
	pad := numWidth(total)
	numStr := fmt.Sprintf("%*d", pad, num)
	return lineNumStyle.Render(numStr) + lineNumSepStyle.Render(" │ ")
}

func lineNumBlank(total int) string {
	pad := numWidth(total)
	return strings.Repeat(" ", pad) + "   "
}
