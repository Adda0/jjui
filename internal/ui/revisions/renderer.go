package revisions

import (
	"jjui/internal/jj"
	"strings"

	"jjui/internal/ui/common"

	tea "github.com/charmbracelet/bubbletea"
)

type SegmentedRenderer struct {
	Palette             common.Palette
	HighlightedRevision string
	Overlay             tea.Model
	op                  common.Operation
}

func (s SegmentedRenderer) Render(context *jj.TreeRow) {
	commit := context.Commit
	highlighted := commit.ChangeIdShort == s.HighlightedRevision
	if (s.op == common.RebaseBranch || s.op == common.RebaseRevision) && highlighted {
		context.Before = common.DropStyle.Render("<< here >>")
	}

	style := s.Palette.Normal
	if highlighted {
		style = s.Palette.Selected
	}
	if commit.Immutable || commit.IsRoot() {
		context.Glyph = style.Render("◆")
	} else if commit.IsWorkingCopy {
		context.Glyph = style.Render("@")
	} else if commit.Conflict {
		context.Glyph = style.Render("×")
	} else {
		context.Glyph = style.Render("○")
	}

	var w strings.Builder
	w.WriteString(s.Palette.CommitShortStyle.Render(commit.ChangeIdShort))
	w.WriteString(s.Palette.CommitIdRestStyle.Render(commit.ChangeId[len(commit.ChangeIdShort):]))
	w.WriteString(" ")

	if commit.IsRoot() {
		w.WriteString(s.Palette.Empty.Render("root()"))
		w.Write([]byte{'\n'})
		context.Content = w.String()
		return
	}

	w.WriteString(s.Palette.AuthorStyle.Render(commit.Author))
	w.WriteString(" ")

	w.WriteString(s.Palette.TimestampStyle.Render(commit.Timestamp))
	w.WriteString(" ")

	w.WriteString(s.Palette.BookmarksStyle.Render(strings.Join(commit.Bookmarks, " ")))

	if commit.Conflict {
		w.WriteString(" ")
		w.WriteString(s.Palette.ConflictStyle.Render("conflict"))
	}
	w.Write([]byte{'\n'})
	if s.op == common.EditDescription && highlighted {
		w.WriteString(s.Overlay.View())
		w.Write([]byte{'\n'})
		context.Content = w.String()
		return
	}
	if commit.Empty {
		w.WriteString(s.Palette.Empty.Render("(empty)"))
		w.WriteString(" ")
	}
	if commit.Description == "" {
		if commit.Empty {
			w.WriteString(s.Palette.Empty.Render("(no description set)"))
		} else {
			w.WriteString(s.Palette.NonEmpty.Render("(no description set)"))
		}
	} else {
		w.WriteString(s.Palette.Normal.Render(commit.Description))
	}
	w.Write([]byte{'\n'})
	if s.Overlay != nil && highlighted {
		w.WriteString(s.Overlay.View())
		w.Write([]byte{'\n'})
	}
	context.Content = w.String()

	if context.EdgeType == jj.IndirectEdge {
		context.ElidedRevision = s.Palette.CommitIdRestStyle.Render("~  (elided revisions)")
	}
}
