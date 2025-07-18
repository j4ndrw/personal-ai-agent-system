package ui

import (
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) RenderMessages() error {
	renderedMessages, err := glamour.Render(
		lipgloss.
			NewStyle().
			Width(m.viewport.Width).
			Render(strings.Join(m.messages, "\n")),
		"dark",
	)
	if err != nil {
		return err
	}
	m.viewport.SetContent(renderedMessages)
	m.viewport.GotoBottom()
	return nil
}
