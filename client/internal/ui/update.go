package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/j4ndrw/personal-ai-agent-system/client/internal/agent"
)

func (m *Model) ErrorUpdate(msg ErrMsg) (tea.Model, tea.Cmd) {
	m.state.Err = msg
	return m, nil
}

func (m *Model) ReceiveStreamChunkTickUpdate(msg agent.ReceiveStreamChunkTickMsg) (tea.Model, tea.Cmd) {
	cmd := m.ReceiveStreamChunkTickHandler(msg)
	return m, cmd
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
	cmd := m.QuitKeyHandler()
	return m, cmd
}

func (m *Model) ChatMessageSendUpdate() (tea.Model, tea.Cmd) {
	cmd, err := m.ChatMessageSendHandler()
	if err != nil {
		return m.ErrorUpdate(err)
	}
	return m, cmd
}

func (m *Model) ScrollUpUpdate() (tea.Model, tea.Cmd) {
	cmd, err := m.ScrollUpHandler()
	if err != nil {
		return m.ErrorUpdate(err)
	}
	return m, cmd
}

func (m *Model) ScrollDownUpdate() (tea.Model, tea.Cmd) {
	cmd, err := m.ScrollDownHandler()
	if err != nil {
		return m.ErrorUpdate(err)
	}
	return m, cmd
}

func (m *Model) YankUpdate() (tea.Model, tea.Cmd) {
	cmd, err := m.YankHandler()
	if err != nil {
		return m.ErrorUpdate(err)
	}
	return m, cmd
}

func (m *Model) InspectThoughtsUpdate() (tea.Model, tea.Cmd) {
	cmd, err := m.InspectThoughtsHandler()
	if err != nil {
		return m.ErrorUpdate(err)
	}
	return m, cmd
}

func (m *Model) InspectAnswersUpdate() (tea.Model, tea.Cmd) {
	cmd, err := m.InspectAnswersHandler()
	if err != nil {
		return m.ErrorUpdate(err)
	}
	return m, cmd
}

func (m *Model) InspectToolCallsUpdate() (tea.Model, tea.Cmd) {
	cmd, err := m.InspectToolCallsHandler()
	if err != nil {
		return m.ErrorUpdate(err)
	}
	return m, cmd
}

func (m *Model) ToNormalModeUpdate() (tea.Model, tea.Cmd) {
	cmd, err := m.ToNormalModeHandler()
	if err != nil {
		return m.ErrorUpdate(err)
	}
	return m, cmd
}

func (m *Model) ToInsertModeUpdate() (tea.Model, tea.Cmd) {
	cmd, err := m.ToInsertModeHandler()
	if err != nil {
		return m.ErrorUpdate(err)
	}
	return m, cmd
}

func (m *Model) ScrollToTopUpdate() (tea.Model, tea.Cmd) {
	cmd, err := m.ScrollToTopHandler()
	if err != nil {
		return m.ErrorUpdate(err)
	}
	return m, cmd
}

func (m *Model) ScrollToBottomUpdate() (tea.Model, tea.Cmd) {
	cmd, err := m.ScrollToBottomHandler()
	if err != nil {
		return m.ErrorUpdate(err)
	}
	return m, cmd
}

func (m *Model) CycleChatModeUpdate() (tea.Model, tea.Cmd) {
	cmd, err := m.CycleChatModeHandler()
	if err != nil {
		return m.ErrorUpdate(err)
	}
	return m, cmd
}
