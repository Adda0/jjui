package graph

import (
	"bytes"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/idursun/jjui/internal/config"
	"github.com/idursun/jjui/internal/ui/common"
	"strings"
)

type GraphWriter struct {
	buffer    bytes.Buffer
	lineCount int
	Width     int
}

func (w *GraphWriter) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	w.lineCount += bytes.Count(p, []byte("\n"))
	return w.buffer.Write(p)
}

func (w *GraphWriter) LineCount() int {
	return w.lineCount
}

func (w *GraphWriter) String(start, end int) string {
	lines := strings.Split(w.buffer.String(), "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	if start < 0 {
		start = 0
	}
	if end < start {
		end = start
	}
	for end > len(lines) {
		lines = append(lines, "")
	}
	return strings.Join(lines[start:end], "\n")
}

func (w *GraphWriter) Reset() {
	w.buffer.Reset()
	w.lineCount = 0
}

func (w *GraphWriter) RenderRow(row GraphRow, renderer RowDecorator, highlighted bool) {
	// will render by extending the previous connections
	before := renderer.RenderBefore(row.Commit)
	if before != "" {
		extended := GraphRowLine{}
		if row.Previous != nil {
			extended = row.Previous.Last(Highlightable).Extend(row.Indent)
		}
		lines := strings.Split(before, "\n")
		for _, line := range lines {
			for _, segment := range extended.Segments {
				fmt.Fprint(&w.buffer, segment.String())
			}
			fmt.Fprintln(&w.buffer, line)
		}
	}

	highlightColor := lipgloss.AdaptiveColor{
		Light: config.Current.UI.HighlightLight,
		Dark:  config.Current.UI.HighlightDark,
	}

	highlightSeq := lipgloss.ColorProfile().FromColor(highlightColor).Sequence(true)
	var lastLine *GraphRowLine
	for segmentedLine := range row.SegmentLinesIter(Including(Highlightable)) {
		lastLine = segmentedLine
		lw := strings.Builder{}
		for i, segment := range segmentedLine.Segments {
			if i == segmentedLine.ChangeIdIdx {
				if decoration := renderer.RenderBeforeChangeId(); decoration != "" {
					fmt.Fprint(&lw, decoration, " ")
				}
			}
			if highlighted && i == segmentedLine.CommitIdIdx {
				if decoration := renderer.RenderBeforeCommitId(); decoration != "" {
					fmt.Fprint(&lw, decoration)
				}
			}
			if highlighted {
				fmt.Fprint(&lw, segment.WithBackground(highlightSeq))
			} else {
				fmt.Fprint(&lw, segment.String())
			}
		}
		if segmentedLine.Flags&Revision == Revision && row.IsAffected {
			style := common.DefaultPalette.Dimmed
			if highlighted {
				style = common.DefaultPalette.Dimmed.Background(highlightColor)
			}
			fmt.Fprint(&lw, style.Render(" (affected by last operation)"))
		}
		line := lw.String()
		fmt.Fprint(w, line)
		if highlighted {
			width := lipgloss.Width(line)
			gap := w.Width - width
			if gap > 0 {
				fmt.Fprintf(w, "\033[%sm%s\033[0m", highlightSeq, strings.Repeat(" ", gap))
			}
		}
		fmt.Fprint(w, "\n")
	}

	if row.Commit.IsRoot() {
		return
	}

	afterSection := renderer.RenderAfter(row.Commit)
	if afterSection != "" && lastLine != nil {
		extended := lastLine.Extend(row.Indent)
		lines := strings.Split(afterSection, "\n")
		for _, line := range lines {
			for _, segment := range extended.Segments {
				fmt.Fprint(w, segment.String())
			}
			fmt.Fprintln(w, line)
		}
	}

	for segmentedLine := range row.SegmentLinesIter(Excluding(Highlightable)) {
		for _, segment := range segmentedLine.Segments {
			fmt.Fprint(w, segment.String())
		}
		fmt.Fprint(w, "\n")
	}
}
