package ui

import (
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) RenderMessagesUtil() error {
	renderedMessages, err := glamour.Render(
		lipgloss.
			NewStyle().
			Width(m.viewport.Width).
			Render(strings.Join(m.state.messages, "\n")),
		"dark",
	)
	if err != nil {
		return err
	}
	m.viewport.SetContent(renderedMessages)
	m.viewport.GotoBottom()
	return nil
}
