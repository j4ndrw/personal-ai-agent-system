package ui

import (
	"fmt"
	"strings"
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

	message := ""
	for _, line := range strings.Split(prompt, "\n") {
		message += "> " + line + "\n"
	}
	m.state.UserMessages = append(m.state.UserMessages, message)
	m.state.AgentAnswers = append(m.state.AgentAnswers, "")
	m.state.AgentThoughts = append(m.state.AgentThoughts, "")
	m.state.AgentToolCalls = append(m.state.AgentToolCalls, "")

	err := m.RenderMessagesUtil()
	if err != nil {
		return nil, err
	}
	m.textarea.Reset()

	msg, err := agent.OpenStream(prompt)
	if err != nil {
		return nil, err
	}

	m.ResetAgentState()
	m.state.Waiting = true
	m.state.Async.ReadChunk.Data = &state.ReadChunkData{
		Result: agent.ReceiveStreamChunkMsg{},
		Err:    nil,
		Phase:  async.ReadyAsyncResultState,
	}
	return m.ToCmd(msg), nil
}

func (m *Model) ResetAgentState() {
	m.state.Async.ReadChunk.Data = nil
	m.state.Agent.Token = ""
	m.state.Agent.ToolCall = ""
	m.state.Waiting = false
	m.state.Agent.ProcessedChunkIds = []string{}
	m.state.Agent.ChunkId = ""
}

func (m *Model) ReceiveStreamChunkTickHandler(msg agent.ReceiveStreamChunkTickMsg) tea.Cmd {
	if m.state.Async.ReadChunk.Data == nil {
		return nil
	}

	return tea.Tick(time.Millisecond, func(t time.Time) tea.Msg {
		return agent.ReceiveStreamChunkMsg{
			AgentChunk: msg.AgentChunk,
			Response:   msg.Response,
			Time:       t,
		}
	})
}

func (m *Model) ReceiveStreamChunkHandler(msg agent.ReceiveStreamChunkMsg) (tea.Cmd, error) {
	if m.state.Async.ReadChunk.Data == nil {
		return nil, nil
	}

	toCmd := func(msg agent.ReceiveStreamChunkMsg) tea.Cmd {
		return m.ToCmd(agent.ReceiveStreamChunkTickMsg{
			AgentChunk: msg.AgentChunk,
			Response:   msg.Response,
			Time:       msg.Time,
		})
	}

	switch m.state.Async.ReadChunk.Data.Phase {
	case async.ReadyAsyncResultState:
		{
			m.state.Async.ReadChunk.Data.Phase = async.PendingAsyncResultState
			go agent.ReadChunk(msg, m.state.Async.ReadChunk.Data)
			return toCmd(msg), nil
		}

	case async.PendingAsyncResultState:
		{
			return toCmd(msg), nil
		}

	default:
		{
			recvMsg := m.state.Async.ReadChunk.Data.Result.(agent.ReceiveStreamChunkMsg)
			err := m.state.Async.ReadChunk.Data.Err

			if err != nil || recvMsg.AgentChunk.Id == "" {
				m.ResetAgentState()
				return toCmd(msg), err
			}

			agent.MapChunk(
				&m.state.Agent.ChunkId,
				&m.state.Agent.Token,
				&m.state.Agent.ToolCall,
				&m.state.Agent.Thinking,
			)(recvMsg.AgentChunk)

			sinkMap := map[string]func() (*[]string, string){
				"thoughts": func() (*[]string, string) {
					return &m.state.AgentThoughts, m.state.Agent.Token
				},
				"answers": func() (*[]string, string) {
					return &m.state.AgentAnswers, m.state.Agent.Token
				},
				"toolcalls": func() (*[]string, string) {
					return &m.state.AgentToolCalls, m.state.Agent.ToolCall
				},
			}
			for k, _ := range sinkMap {
				thoughts := k == "thoughts" && m.state.Agent.Thinking && m.state.Agent.Token != ""
				answers := k == "answers" && !m.state.Agent.Thinking && m.state.Agent.Token != ""
				toolcalls := k == "toolcalls" && m.state.Agent.ToolCall != ""
				if thoughts || answers || toolcalls {
					sink, chunk := sinkMap[k]()
					if err := agent.ProcessChunk(
						sink,
						chunk,
						m.state.Agent.ChunkId,
						&m.state.Agent.ProcessedChunkIds,
						m.RenderMessagesUtil,
					); err != nil {
						m.ResetAgentState()
						return toCmd(msg), err
					}
				}
			}

			m.state.Async.ReadChunk.Data.Phase = async.ReadyAsyncResultState
			return toCmd(recvMsg), nil
		}
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
