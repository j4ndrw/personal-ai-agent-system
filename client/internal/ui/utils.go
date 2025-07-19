package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) RenderMessagesUtil() error {
	renderedMessages, err := glamour.Render(
		lipgloss.
			NewStyle().
			Width(m.viewport.Width).
			Render(strings.Join(m.state.Messages, "\n")),
		"dark",
	)
	if err != nil {
		return err
	}
	m.viewport.SetContent(renderedMessages)
	m.viewport.GotoBottom()
	return nil
}

func (m *Model) ToCmd (msg tea.Msg) tea.Cmd {
	return func () tea.Msg {
		return msg
	}
}
