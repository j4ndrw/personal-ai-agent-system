package ui

import (
	"fmt"
	"io"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/j4ndrw/personal-ai-agent-system/client/internal/agent"
	"github.com/j4ndrw/personal-ai-agent-system/client/internal/async"
	"github.com/j4ndrw/personal-ai-agent-system/client/internal/state"
)

func (m *Model) WindowSizeHandler(msg tea.WindowSizeMsg) error {
	m.textarea.SetWidth(msg.Width)
	m.viewport.Width = msg.Width
	m.viewport.Height = func() int {
		diff := msg.Height - lipgloss.Height(Gap)
		if m.state.Waiting {
			return diff - m.spinner.Style.GetHeight()
		}
		return diff - m.textarea.Height()
	}()

	if len(m.state.Messages) > 0 {
		err := m.RenderMessagesUtil()
		if err != nil {
			return err
		}
		return nil
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
	m.state.Messages = append(m.state.Messages, message)
	err := m.RenderMessagesUtil()
	if err != nil {
		return nil, err
	}
	m.textarea.Reset()

	recvMsg, err := agent.OpenStream(prompt)
	if err != nil {
		return nil, err
	}

	m.state.Waiting = true
	m.state.Async.ReadChunk = &state.ReadChunk{
		Result: nil,
		Err:    nil,
		Phase:  async.ReadyAsyncResultState,
	}
	return func() tea.Msg {
		return recvMsg
	}, nil
}

func (m *Model) ProcessAgentChunk(msg *agent.ReceiveStreamChunkMsg) error {
	if msg == nil {
		return agent.ProcessToolCalls(
			&m.state.Messages,
			&m.state.Agent.ToolCalls,
			m.RenderMessagesUtil,
		)
	}
	return agent.ProcessAnswerChunk(
		&m.state.Messages,
		m.state.Agent.Token,
		m.RenderMessagesUtil,
	)

}

func (m *Model) ResetAgentState(msg *agent.ReceiveStreamChunkMsg) {
	if msg == nil {
		m.state.Agent.Token = ""
		m.state.Agent.ToolCalls = []string{}
		m.state.Waiting = false
		m.state.Async.ReadChunk = nil
	}
}

func (m *Model) ReceiveStreamChunkTickHandler(msg agent.ReceiveStreamChunkTickMsg) tea.Cmd {
	return tea.Tick(time.Millisecond, func(t time.Time) tea.Msg {
		return agent.ReceiveStreamChunkMsg{
			AgentChunk: msg.AgentChunk,
			Response:   msg.Response,
			Time:       t,
		}
	})
}

func (m *Model) ReceiveStreamChunkHandler(msg agent.ReceiveStreamChunkMsg) (tea.Cmd, error) {
	if m.state.Async.ReadChunk == nil {
		return nil, nil
	}

	toCmd := func(msg agent.ReceiveStreamChunkMsg) tea.Cmd {
		return m.ToCmd(agent.ReceiveStreamChunkTickMsg{
			AgentChunk: msg.AgentChunk,
			Response:   msg.Response,
			Time:       msg.Time,
		})
	}

	switch m.state.Async.ReadChunk.Phase {
	case async.ReadyAsyncResultState:
		m.state.Async.ReadChunk.Phase = async.PendingAsyncResultState
		go agent.ReadChunk(&msg, m.state.Async.ReadChunk)
		return toCmd(msg), nil
	case async.DoneAsyncResultState:
		err := m.state.Async.ReadChunk.Err
		if err == io.EOF {
			m.state.Async.ReadChunk = nil
			m.state.Waiting = false
			return toCmd(msg), nil
		}
		if err != nil {
			m.state.Async.ReadChunk = nil
			m.state.Waiting = false
			return toCmd(msg), err
		}

		recvMsg := m.state.Async.ReadChunk.Result.(*agent.ReceiveStreamChunkMsg)
		if recvMsg != nil {
			agent.MapChunk(
				&m.state.Agent.Token,
				&m.state.Agent.ToolCalls,
				&m.state.Agent.Thinking,
			)(recvMsg.AgentChunk)
		}

		err = m.ProcessAgentChunk(recvMsg)
		if err != nil {
			m.state.Waiting = false
			m.state.Async.ReadChunk = nil
			return toCmd(msg), err
		}

		m.ResetAgentState(recvMsg)

		if recvMsg == nil {
			m.state.Waiting = false
			return toCmd(msg), nil
		}

		m.state.Async.ReadChunk.Phase = async.ReadyAsyncResultState
		return toCmd(*recvMsg), nil
	default:
		return toCmd(msg), nil
	}
}

func (m *Model) ScrollUpHandler() (tea.Cmd, error) {
	m.viewport.ScrollUp(ScrollSize)
	return nil, nil
}

func (m *Model) ScrollDownHandler() (tea.Cmd, error) {
	m.viewport.ScrollDown(ScrollSize)
	return nil, nil
}

func (m *Model) YankHandler() (tea.Cmd, error) {
	messages := m.GetUnstyledMessagesUtil()
	err := clipboard.WriteAll(messages)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
