package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/j4ndrw/personal-ai-agent-system/client/internal/state"
)

func (m *Model) GetUnstyledMessagesUtil() string {
	sb := ""
	for i, userMessage := range m.state.UserMessages {
		sb += userMessage + "\n"

		sink := func() []string {
			switch m.state.AgentMessageToShow {
			case state.AgentMessageShowAnswers:
				return m.state.AgentAnswers
			case state.AgentMessageShowThoughts:
				return m.state.AgentThoughts
			default:
				return m.state.AgentToolCalls
			}
		}()

		sb += sink[i] + "\n\n---\n"
	}
	return sb
}

func (m *Model) GetFullUnstyledMessagesUtil() string {
	sb := ""
	for i, userMessage := range m.state.UserMessages {
		sb += userMessage + "\n"
		sb += m.state.AgentThoughts[i] + "\n"
		sb += m.state.AgentAnswers[i] + "\n"
		sb += m.state.AgentToolCalls[i] + "\n"
		sb += "\n---\n"
	}
	return sb
}
func (m *Model) GetRenderedMessagesUtil() (string, error) {
	renderedMessages, err := m.markdownRenderer.Render(
		lipgloss.
			NewStyle().
			Width(m.viewport.Width).
			Render(m.GetUnstyledMessagesUtil()),
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

func (m *Model) ToCmd(msg tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}
