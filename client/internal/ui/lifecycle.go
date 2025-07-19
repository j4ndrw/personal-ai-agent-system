package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/j4ndrw/personal-ai-agent-system/client/internal/agent"
)

func (m *Model) ErrorUpdate(msg ErrMsg) (tea.Model, tea.Cmd) {
	m.state.err = msg
	return m, nil
}

func (m *Model) ReceiveStreamChunkUpdate(msg agent.ReceiveStreamChunkMsg) (tea.Model, tea.Cmd) {
	cmd, err := m.ReceiveStreamChunkHandler(msg)
	if cmd != nil {
		return m, cmd
	}
	if err != nil {
		return m.ErrorUpdate(err)
	}
	return m, nil
}

func (m *Model) WindowSizeUpdate(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	err := m.WindowSizeHandler(msg)
	if err != nil {
		return m.ErrorUpdate(err)
	}
	return m, nil
}

func (m *Model) QuitKeyUpdate() (tea.Model, tea.Cmd) {
	m.QuitKeyHandler()
	return m, tea.Quit
}

func (m *Model) ChatMessageSendUpdate() (tea.Model, tea.Cmd) {
	cmd, err := m.ChatMessageSendHandler()
	if err != nil {
		return m.ErrorUpdate(err)
	}
	return m, cmd
}
