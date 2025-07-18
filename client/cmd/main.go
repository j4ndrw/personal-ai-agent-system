package main

import (
	"fmt"
	"log"
	"strings"

	// bubbles "github.com/charmbracelet/bubbles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/j4ndrw/personal-ai-agent-system/client/internal/agent"

	// glamour "github.com/charmbracelet/glamour"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
)

const gap = "\n"

type errMsg error

type model struct {
	viewport            viewport.Model
	messages            []string
	textarea            textarea.Model
	agentMessageChannel chan string
	err                 error
}

func initialModel() model {
	ta := textarea.New()
	ta.Placeholder = "Chat with the agent..."
	ta.Focus()

	ta.Prompt = "┃ "

	ta.SetWidth(30)
	ta.SetHeight(3)

	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(30, 5)
	ta.KeyMap.InsertNewline.SetEnabled(false)

	return model{
		textarea:            ta,
		messages:            []string{},
		agentMessageChannel: nil,
		viewport:            vp,
		err:                 nil,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) ListenForAgentMessages() {
	if m.agentMessageChannel == nil {
		return
	}

	prefix := "# Agent\n"
	if len(m.messages) > 0 && m.messages[len(m.messages)-1] != prefix {
		m.messages = append(m.messages, prefix)
		m.viewport.SetContent(
			lipgloss.
				NewStyle().
				Width(m.viewport.Width).
				Render(strings.Join(m.messages, "\n")),
		)
		m.viewport.GotoBottom()
	}

	go func(m *model) {
		agentMessage, done := <-m.agentMessageChannel
		renderedAgentMessage, err := glamour.Render(agentMessage, "dark")
		if err != nil {
			return
		}

		m.messages[len(m.messages)-1] = renderedAgentMessage
		if done {
			m.agentMessageChannel = nil
		}
	}(&m)

	m.viewport.SetContent(
		lipgloss.
			NewStyle().
			Width(m.viewport.Width).
			Render(strings.Join(m.messages, "\n")),
	)
	m.viewport.GotoBottom()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.textarea.SetWidth(msg.Width)
		m.viewport.Height = msg.Height - m.textarea.Height() - lipgloss.Height(gap)

		if len(m.messages) > 0 {
			m.viewport.SetContent(
				lipgloss.
					NewStyle().
					Width(m.viewport.Width).
					Render(strings.Join(m.messages, "\n")),
			)
		}
		m.viewport.GotoBottom()
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case tea.KeyEnter:
			prompt := m.textarea.Value()
			message, err := glamour.Render("# User\n"+prompt, "dark")
			if err != nil {
				m.err = err
				return m, nil
			}
			m.messages = append(m.messages, message)
			m.viewport.SetContent(
				lipgloss.
					NewStyle().
					Width(m.viewport.Width).
					Render(strings.Join(m.messages, "\n")),
			)
			m.textarea.Reset()
			m.viewport.GotoBottom()

			if m.agentMessageChannel == nil {
				m.agentMessageChannel = make(chan string)
			}
			var thinking bool = false
			go agent.Chat(
				prompt,
				func(chunk agent.AgentChunk) {
					if chunk.Answer == nil && chunk.ToolCall == nil {
						return
					}

					if chunk.ToolCall != nil {
						m.agentMessageChannel <- "`tool:`" + chunk.ToolCall.ToolCall + "\n```json\n" + chunk.ToolCall.JSONResult + "```\n\n"
						return
					}

					if thinking != chunk.Answer.Thinking == true {
						thinking = chunk.Answer.Thinking
						if thinking {
							m.agentMessageChannel <- "**Thinking**\n\n"
						} else {
							m.agentMessageChannel <- "**Done thinking**\n\n"
						}
					}

					m.agentMessageChannel <- chunk.Answer.Content
				},
				func() {
					close(m.agentMessageChannel)
				})
		}

	case errMsg:
		m.err = msg
		return m, nil
	}

	m.ListenForAgentMessages()
	return m, tea.Batch(tiCmd, vpCmd)
}

func (m model) View() string {
	return fmt.Sprintf(
		"%s%s%s",
		m.viewport.View(),
		gap,
		m.textarea.View(),
	)
}

func main() {
	p := tea.NewProgram(initialModel())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
