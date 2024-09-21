package revisions

import (
	"fmt"
	"os"
	"strings"

	"jjui/internal/dag"
	"jjui/internal/jj"
	"jjui/internal/ui/bookmark"
	"jjui/internal/ui/common"
	"jjui/internal/ui/describe"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type mode int

const (
	normalMode mode = iota
	moveMode
)

type Model struct {
	rows       []dag.GraphRow
	mode       mode
	draggedRow int
	cursor     int
	width      int
	help       help.Model
	describe   tea.Model
	bookmarks  tea.Model
	keymap     keymap
}

type keymap struct {
	current  rune
	bindings map[rune][]key.Binding
}

func newKeyMap() keymap {
	bindings := make(map[rune][]key.Binding)
	bindings[' '] = []key.Binding{
		key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j", "down")),
		key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k", "up")),
		key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "set description")),
		key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
	}
	bindings['r'] = []key.Binding{
		key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "rebase revision")),
		key.NewBinding(key.WithKeys("b"), key.WithHelp("b", "rebase branch")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "rebase apply")),
		key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
	}
	bindings['b'] = []key.Binding{
		key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "bookmark set")),
		key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "bookmark delete")),
		key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
	}
	bindings['g'] = []key.Binding{
		key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "git push")),
		key.NewBinding(key.WithKeys("f"), key.WithHelp("d", "git fetch")),
		key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
	}

	return keymap{
		current:  ' ',
		bindings: bindings,
	}
}

func (k keymap) ShortHelp() []key.Binding {
	return k.bindings[k.current]
}

func (k keymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}

func (m Model) selectedRevision() *jj.Commit {
	return m.rows[m.cursor].Commit
}

func (m Model) Init() tea.Cmd {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return nil
	}
	return common.FetchRevisions(dir)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.describe != nil {
		switch msg.(type) {
		case common.CloseViewMsg:
			m.describe = nil
			return m, nil
		}

		var cmd tea.Cmd
		m.describe, cmd = m.describe.Update(msg)
		return m, cmd
	}

	if m.bookmarks != nil {
		switch msg.(type) {
		case common.CloseViewMsg:
			m.bookmarks = nil
			return m, nil
		}

		var cmd tea.Cmd
		m.bookmarks, cmd = m.bookmarks.Update(msg)
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "g":
			m.keymap.current = 'g'
			return m, nil
		case "r":
			m.keymap.current = 'r'
			return m, nil
		case "b":
			m.keymap.current = 'b'
			return m, common.FetchBookmarks
		case "d":
			return m, common.ShowDescribe(m.selectedRevision())
		case "down", "j":
			if m.cursor < len(m.rows)-1 {
				m.cursor++
			}
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "esc":
			if m.keymap.current != ' ' {
				m.keymap.current = ' '
				return m, nil
			}
			m.draggedRow = -1
			m.mode = normalMode
		case " ":
			if m.mode == normalMode {
				m.mode = moveMode
				m.draggedRow = m.cursor
			} else {
				m.mode = normalMode
				m.draggedRow = -1
			}
		case "enter":
			if m.mode == moveMode {
				m.mode = normalMode
				fromRevision := m.rows[m.draggedRow].Commit.ChangeIdShort
				toRevision := m.rows[m.cursor].Commit.ChangeIdShort
				m.draggedRow = -1
				return m, common.RebaseCommand(fromRevision, toRevision)
			}
		default:
			return m, nil
		}
	case common.UpdateRevisionsMsg:
		m.rows = []dag.GraphRow(msg)
	case common.UpdateBookmarksMsg:
		m.bookmarks = bookmark.New(m.selectedRevision().ChangeId, msg, m.width)
	case common.ShowDescribeViewMsg:
		m.describe = describe.New(msg.ChangeId, msg.Description, m.width)
		return m, m.describe.Init()
	case common.CloseViewMsg:
		m.describe = nil
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
	}
	return m, nil
}

func (m Model) View() string {
	var items strings.Builder

	for i := 0; i < len(m.rows); i++ {
		row := &m.rows[i]
		switch m.mode {
		case moveMode:
			if i == m.cursor {
				indent := strings.Repeat("│ ", row.Level)
				items.WriteString(indent)
				items.WriteString(common.DropStyle.Render("<< here >>"))
				items.WriteString("\n")
			}
			DefaultRenderer(&items, row, common.DefaultPalette, i == m.draggedRow)
		case normalMode:
			DefaultRenderer(&items, row, common.DefaultPalette, i == m.cursor)
			if m.describe != nil && m.cursor == i {
				items.WriteString(m.describe.View())
				items.WriteString("\n")
			}
			if m.bookmarks != nil && m.cursor == i {
				items.WriteString(m.bookmarks.View())
				items.WriteString("\n")
			}
		}
	}
	items.WriteString("\n")
	items.WriteString(m.help.View(m.keymap))
	items.WriteString("\n")
	if m.mode == moveMode {
		if m.cursor == len(m.rows) {
			items.WriteString("jj rebase -r " + m.rows[m.draggedRow].Commit.ChangeIdShort + " --insert-before " + m.rows[len(m.rows)-1].Commit.ChangeIdShort + "\n")
		} else {
			items.WriteString("jj rebase -r " + m.rows[m.draggedRow].Commit.ChangeIdShort + " -d " + m.rows[m.cursor].Commit.ChangeIdShort + "\n")
		}
	}
	return items.String()
}

func New() tea.Model {
	help := help.New()
	help.Styles.ShortKey = common.DefaultPalette.CommitShortStyle
	help.Styles.ShortDesc = common.DefaultPalette.CommitIdRestStyle
	return Model{
		rows:       []dag.GraphRow{},
		draggedRow: -1,
		mode:       normalMode,
		cursor:     0,
		width:      20,
		keymap:     newKeyMap(),
		describe:   nil,
		help:       help,
	}
}
