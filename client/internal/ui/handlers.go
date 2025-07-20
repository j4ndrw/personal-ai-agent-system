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

	if len(m.state.UserMessages) > 0 {
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
	if prompt == "" {
		return nil, nil
	}

	message := "> " + prompt + "\n"
	m.state.UserMessages = append(m.state.UserMessages, message)
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

func (m *Model) ResetAgentState() {
	m.state.Async.ReadChunk = nil
	m.state.Agent.Token = ""
	m.state.Waiting = false
	m.state.Async.ReadChunk = nil
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
			m.ResetAgentState()
			return toCmd(msg), nil
		}
		if err != nil {
			m.ResetAgentState()
			return toCmd(msg), err
		}
		if m.state.Async.ReadChunk.Result == nil {
			m.ResetAgentState()
			return toCmd(msg), err
		}

		recvMsg := m.state.Async.ReadChunk.Result.(*agent.ReceiveStreamChunkMsg)
		agent.MapChunk(
			&m.state.Agent.Token,
			&m.state.Agent.ToolCall,
			&m.state.Agent.Thinking,
		)(recvMsg.AgentChunk)

		sinkMap := map[string]func() (*[]string, string, bool){
			"thoughts": func() (*[]string, string, bool) {
				return &m.state.AgentThoughts, m.state.Agent.Token, false
			},
			"answers": func() (*[]string, string, bool) {
				return &m.state.AgentAnswers, m.state.Agent.Token, false
			},
			"toolcalls": func() (*[]string, string, bool) {
				return &m.state.AgentToolCalls, m.state.Agent.ToolCall, true
			},
		}
		for k, _ := range sinkMap {
			if (k == "thoughts" && m.state.Agent.Thinking) ||
				(k == "answers" && !m.state.Agent.Thinking) ||
				(k == "toolcalls" && m.state.Agent.ToolCall != "") {
				sink, chunk, idempotent := sinkMap[k]()
				err = agent.ProcessChunk(
					sink,
					chunk,
					m.RenderMessagesUtil,
					idempotent,
				)
				if err != nil {
					m.ResetAgentState()
					return toCmd(msg), err
				}
			}
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
	messages := m.GetFullUnstyledMessagesUtil()
	err := clipboard.WriteAll(messages)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (m *Model) InspectThoughtsHandler() (tea.Cmd, error) {
	m.state.AgentMessageToShow = state.AgentMessageShowThoughts
	err := m.RenderMessagesUtil()
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (m *Model) InspectAnswersHandler() (tea.Cmd, error) {
	m.state.AgentMessageToShow = state.AgentMessageShowAnswers
	err := m.RenderMessagesUtil()
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (m *Model) InspectToolCallsHandler() (tea.Cmd, error) {
	m.state.AgentMessageToShow = state.AgentMessageShowToolCalls
	err := m.RenderMessagesUtil()
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (m *Model) ToNormalModeHandler() (tea.Cmd, error) {
	m.state.Mode = state.NormalMode
	m.textarea.Blur()
	return nil, nil
}

func (m *Model) ToInsertModeHandler() (tea.Cmd, error) {
	if m.state.Waiting {
		return nil, nil
	}
	m.state.Mode = state.InsertMode
	return m.textarea.Focus(), nil
}

func (m *Model) NewLineHandler() (tea.Cmd, error) {
	m.textarea.KeyMap.InsertNewline.SetEnabled(true)
	return nil, nil
}

func (m *Model) ScrollToTopHandler() (tea.Cmd, error) {
	m.viewport.GotoTop()
	return nil, nil
}

func (m *Model) ScrollToBottomHandler() (tea.Cmd, error) {
	m.viewport.GotoBottom()
	return nil, nil
}
