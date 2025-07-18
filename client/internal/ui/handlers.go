package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/j4ndrw/personal-ai-agent-system/client/internal/agent"
)

func (m *Model) WindowSizeHandler(msg tea.WindowSizeMsg) error {
	m.textarea.SetWidth(msg.Width)
	m.viewport.Width = msg.Width
	m.viewport.Height = msg.Height - m.textarea.Height() - lipgloss.Height(Gap)

	if len(m.messages) > 0 {
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
	}
	m.viewport.GotoBottom()
	return nil
}

func (m *Model) QuitKeyHandler() {
	fmt.Println(m.textarea.Value())
}

func (m *Model) ChatMessageSendHandler() (tea.Cmd, error) {
	prompt := m.textarea.Value()
	message := "> " + prompt + "\n\n"
	m.messages = append(m.messages, message)
	renderedMessages, err := glamour.Render(
		lipgloss.
			NewStyle().
			Width(m.viewport.Width).
			Render(strings.Join(m.messages, "\n")),
		"dark",
	)
	if err != nil {
		return nil, err
	}
	m.viewport.SetContent(renderedMessages)
	m.textarea.Reset()
	m.viewport.GotoBottom()

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
		agent.MapChunk(&m.agentMessageChunk, &m.toolCalls, &m.agentThinking),
	)

	if err != nil {
		return nil, err
	}

	if recvMsg == nil {
		err = agent.ProcessChunk(
			&m.messages,
			&m.toolCalls,
			"",
			m.RenderMessages,
		)
		if err != nil {
			return nil, err
		}

		m.agentMessageChunk = ""
		m.toolCalls = []string{}
		return nil, nil
	}

	err = agent.ProcessChunk(
		&m.messages,
		&[]string{},
		m.agentMessageChunk,
		m.RenderMessages,
	)
	if err != nil {
		return nil, err
	}

	return func() tea.Msg {
		return *recvMsg
	}, nil
}
