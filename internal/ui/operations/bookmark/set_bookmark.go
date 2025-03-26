package bookmark

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/idursun/jjui/internal/jj"
	"github.com/idursun/jjui/internal/ui/common"
	"github.com/idursun/jjui/internal/ui/context"
	"github.com/idursun/jjui/internal/ui/operations"
)

type SetBookmarkOperation struct {
	context  context.AppContext
	revision string
	name     textarea.Model
}

func (s SetBookmarkOperation) Init() tea.Cmd {
	return textarea.Blink
}

func (s SetBookmarkOperation) View() string {
	return s.name.View()
}

func (s SetBookmarkOperation) IsFocused() bool {
	return true
}

func (s SetBookmarkOperation) Update(msg tea.Msg) (operations.OperationWithOverlay, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return s, common.Close
		case "enter":
			return s, s.context.RunCommand(jj.BookmarkSet(s.revision, s.name.Value()), common.Close, common.Refresh)
		}
	}
	var cmd tea.Cmd
	s.name, cmd = s.name.Update(msg)
	return s, cmd
}

func (s SetBookmarkOperation) Render() string {
	return s.name.View()
}

func (s SetBookmarkOperation) RenderPosition() operations.RenderPosition {
	return operations.RenderBeforeCommitId
}

func (s SetBookmarkOperation) Name() string {
	return "bookmark"
}

func NewSetBookmarkOperation(context context.AppContext, changeId string) (operations.Operation, tea.Cmd) {
	t := textarea.New()
	t.CharLimit = 120
	t.ShowLineNumbers = false
	t.SetValue("")
	t.SetWidth(20)
	t.SetHeight(1)
	t.Focus()

	op := SetBookmarkOperation{
		name:     t,
		revision: changeId,
		context:  context,
	}
	return op, op.Init()
}
