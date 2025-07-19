package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/j4ndrw/personal-ai-agent-system/client/internal/agent"
)

func (m *Model) WindowSizeHandler(msg tea.WindowSizeMsg) error {
	m.textarea.SetWidth(msg.Width)
	m.viewport.Width = msg.Width
	m.viewport.Height = msg.Height - m.textarea.Height() - lipgloss.Height(Gap)

	if len(m.state.messages) > 0 {
		err := m.RenderMessagesUtil()
		if err != nil {
			return err
		}
	} else {
		m.viewport.GotoBottom()
	}
	return nil
}

func (m *Model) QuitKeyHandler() {
	fmt.Println(m.textarea.Value())
}

func (m *Model) ChatMessageSendHandler() (tea.Cmd, error) {
	prompt := m.textarea.Value()
	message := "> " + prompt + "\n\n"
	m.state.messages = append(m.state.messages, message)
	err := m.RenderMessagesUtil()
	if err != nil {
		return nil, err
	}
	m.textarea.Reset()

	recvMsg, err := agent.OpenStream(prompt)
	if err != nil {
		return nil, err
	}

	return func() tea.Msg {
		return *recvMsg
	}, nil
}

func (m *Model) ReceiveStreamChunkHandler(msg agent.ReceiveStreamChunkMsg) (tea.Cmd, error) {
	recvMsg, err := agent.ReadChunk(
		msg,
		agent.MapChunk(&m.state.agentMessageChunk, &m.state.toolCalls, &m.state.agentThinking),
	)
	if err != nil {
		return nil, err
	}

	err = func(m *Model) error {
		if recvMsg == nil {
			return agent.ProcessToolCalls(
				&m.state.messages,
				&m.state.toolCalls,
				m.RenderMessagesUtil,
			)
		}
		return agent.ProcessAnswerChunk(
			&m.state.messages,
			m.state.agentMessageChunk,
			m.RenderMessagesUtil,
		)
	}(m)
	if err != nil {
		return nil, err
	}

	if recvMsg == nil {
		m.state.agentMessageChunk = ""
		m.state.toolCalls = []string{}
		return nil, nil
	}

	return func() tea.Msg {
		return *recvMsg
	}, nil
}
