package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) GetUnstyledMessagesUtil() string {
	return strings.Join(m.state.Messages, "\n")
}
func (m *Model) GetRenderedMessagesUtil() (string, error) {
	renderedMessages, err := glamour.Render(
		lipgloss.
			NewStyle().
			Width(m.viewport.Width).
			Render(m.GetUnstyledMessagesUtil()),
		"dark",
	)
	if err != nil {
		return "", err
	}
	return renderedMessages, nil
}
func (m *Model) RenderMessagesUtil() error {
	renderedMessages, err := m.GetRenderedMessagesUtil()
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
