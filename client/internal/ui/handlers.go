package ui

import (
	"fmt"
	"log"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/j4ndrw/personal-ai-agent-system/client/internal/agent"
	"github.com/j4ndrw/personal-ai-agent-system/client/internal/async"
	"github.com/j4ndrw/personal-ai-agent-system/client/internal/state"
	"github.com/j4ndrw/personal-ai-agent-system/client/internal/stringtransforms"
)

func (m *Model) WindowSizeHandler(msg tea.WindowSizeMsg) error {
	m.textinput.Width = msg.Width
	m.viewport.Width = msg.Width
	m.viewport.Height = func() int {
		diff := msg.Height - lipgloss.Height(Gap)
		if m.state.Waiting {
			return diff - m.spinner.Style.GetHeight()
		}
		return diff - 1
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
	fmt.Println(m.textinput.Value())
}

func (m *Model) resetAgentState() {
	m.state.Async.ReadChunk.Data = nil
	m.state.Agent.Token = ""
	m.state.Agent.ToolCall = ""
	m.state.Waiting = false
	m.state.Agent.ProcessedChunkIds = []string{}
	m.state.Agent.ChunkId = ""
	m.state.Mode = state.NormalMode
}

func (m *Model) ChatMessageSendHandler() (tea.Cmd, error) {
	m.state.Mode = state.NormalMode

	prompt := m.textinput.Value()
	if prompt == "" {
		return nil, nil
	}

	if m.state.ChatMode == state.AgenticManualChatMode {
		var err error
		m.state.Agent.SelectedAgent, prompt, err = stringtransforms.ExtractAgentAndPrompt(
			prompt,
			(*agent.Agents)(&m.state.Agents),
		)
		if err != nil {
			m.state.Err = err
			log.Fatal(err)
			return nil, nil
		}
	}

	message := stringtransforms.MapUserMessage(prompt)
	m.state.UserMessages = append(m.state.UserMessages, message)
	m.state.AgentAnswers = append(m.state.AgentAnswers, "")
	m.state.AgentThoughts = append(m.state.AgentThoughts, "")
	m.state.AgentToolCalls = append(m.state.AgentToolCalls, "")

	err := m.RenderMessagesUtil()
	if err != nil {
		return nil, err
	}
	m.textinput.Reset()

	endpointMap := map[state.ChatMode]string{
		state.SimpleChatMode:        agent.SimpleEndpoint,
		state.AgenticAutoChatMode:   agent.AgenticAutoEndpoint,
		state.AgenticManualChatMode: agent.AgenticManualEndpoint,
	}
	endpoint := endpointMap[m.state.ChatMode]

	var msg agent.ReceiveStreamChunkMsg
	if m.state.ChatMode == state.AgenticManualChatMode {
		msg, err = agent.OpenAgenticManualStream(prompt, m.state.Agent.SelectedAgent, endpoint)
		if err != nil {
			return nil, err
		}
	} else {
		msg, err = agent.OpenStream(prompt, endpoint)
		if err != nil {
			return nil, err
		}
	}

	m.state.Waiting = true
	m.state.Async.ReadChunk.Data = &state.ReadChunkData{
		Result: agent.ReceiveStreamChunkMsg{},
		Err:    nil,
		Phase:  async.ReadyAsyncResultState,
	}
	return m.ToCmd(msg), nil
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

	if m.state.Async.ReadChunk.Data.Phase == async.ReadyAsyncResultState {
		m.state.Async.ReadChunk.Data.Phase = async.PendingAsyncResultState
		go agent.ReadChunk(msg, m.state.Async.ReadChunk.Data)
	}

	if m.state.Async.ReadChunk.Data.Phase != async.DoneAsyncResultState {
		return toCmd(msg), nil
	}

	recvMsg := m.state.Async.ReadChunk.Data.Result.(agent.ReceiveStreamChunkMsg)
	err := m.state.Async.ReadChunk.Data.Err

	if err != nil || recvMsg.AgentChunk.Id == "" {
		m.resetAgentState()
		return toCmd(msg), err
	}

	err = agent.
		CreateSinkMap(&m.state).
		MapAgentChunk(
			recvMsg.AgentChunk,
			&m.state.Agent,
		).
		ProcessChunk(m.state, agent.ProcessChunk(
			m.state.Agent.ChunkId,
			&m.state.Agent.ProcessedChunkIds,
			m.RenderMessagesUtil,
		))
	if err != nil {
		m.resetAgentState()
		return toCmd(msg), err
	}

	m.state.Async.ReadChunk.Data.Phase = async.ReadyAsyncResultState
	return toCmd(recvMsg), nil

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
	if m.state.ChatMode == state.SimpleChatMode {
		return nil, nil
	}

	m.state.AgentMessageToShow = state.AgentMessageShowThoughts
	err := m.RenderMessagesUtil()
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (m *Model) InspectAnswersHandler() (tea.Cmd, error) {
	if m.state.ChatMode == state.SimpleChatMode {
		return nil, nil
	}

	m.state.AgentMessageToShow = state.AgentMessageShowAnswers
	err := m.RenderMessagesUtil()
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (m *Model) InspectToolCallsHandler() (tea.Cmd, error) {
	if m.state.ChatMode == state.SimpleChatMode {
		return nil, nil
	}

	m.state.AgentMessageToShow = state.AgentMessageShowToolCalls
	err := m.RenderMessagesUtil()
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (m *Model) ToNormalModeHandler() (tea.Cmd, error) {
	m.state.Mode = state.NormalMode
	m.textinput.Blur()
	return nil, nil
}

func (m *Model) ToInsertModeHandler() (tea.Cmd, error) {
	if m.state.Waiting {
		return nil, nil
	}
	m.state.Mode = state.InsertMode
	return m.textinput.Focus(), nil
}

func (m *Model) ScrollToTopHandler() (tea.Cmd, error) {
	m.viewport.GotoTop()
	return nil, nil
}

func (m *Model) ScrollToBottomHandler() (tea.Cmd, error) {
	m.viewport.GotoBottom()
	return nil, nil
}

func (m *Model) CycleChatModeHandler() (tea.Cmd, error) {
	if m.state.Waiting {
		return nil, nil
	}

	switch m.state.ChatMode {
	case state.SimpleChatMode:
		m.state.ChatMode = state.AgenticAutoChatMode
		m.state.Agent.SelectedAgent = ""

		m.textinput.Reset()
		m.textinput.SetSuggestions([]string{})
		break

	case state.AgenticAutoChatMode:
		m.state.ChatMode = state.AgenticManualChatMode
		m.state.Agent.SelectedAgent = ""

		var suggestions []string
		for _, agent := range m.state.Agents {
			suggestions = append(suggestions, fmt.Sprintf("@%s", agent))
		}

		m.textinput.Reset()
		m.textinput.SetSuggestions(suggestions)
		break

	case state.AgenticManualChatMode:
		m.state.ChatMode = state.SimpleChatMode
		m.state.Agent.SelectedAgent = ""
		m.state.AgentMessageToShow = state.AgentMessageShowAnswers

		m.textinput.Reset()
		m.textinput.SetSuggestions([]string{})
		break
	}
	return nil, nil
}
