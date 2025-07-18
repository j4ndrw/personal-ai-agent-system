package main

import (
	"fmt"
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/j4ndrw/personal-ai-agent-system/client/internal/agent"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
)

const gap = "\n"

type errMsg error

type model struct {
	viewport          viewport.Model
	messages          []string
	textarea          textarea.Model
	agentMessageChunk string
	err               error
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
		textarea:          ta,
		messages:          []string{},
		agentMessageChunk: "",
		viewport:          vp,
		err:               nil,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) ListenForAgentMessages() {
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	var thinking bool = false

	switch msg := msg.(type) {
	case agent.ReceiveStreamChunkMsg:
		recvMsg, err := agent.ReadChunk(
			msg,
			agent.MapChunk(&m.agentMessageChunk, &thinking),
		)

		if err != nil {
			m.err = err
			return m, nil
		}

		if recvMsg != nil {
			err := agent.ProcessChunk(
				&m.messages,
				m.agentMessageChunk,
				func() error {
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
					m.viewport.GotoBottom()
					return nil
				})
			if err != nil {
				m.err = err
				return m, nil
			}
			return m, func() tea.Msg {
				return *recvMsg
			}
		} else {
			m.agentMessageChunk = ""
		}

	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.textarea.SetWidth(msg.Width)
		m.viewport.Height = msg.Height - m.textarea.Height() - lipgloss.Height(gap)

		if len(m.messages) > 0 {
			renderedMessages, err := glamour.Render(
				lipgloss.
					NewStyle().
					Width(m.viewport.Width).
					Render(strings.Join(m.messages, "\n")),
				"dark",
			)
			if err != nil {
				m.err = err
				return m, nil
			}
			m.viewport.SetContent(renderedMessages)
		}
		m.viewport.GotoBottom()
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case tea.KeyEnter:
			prompt := m.textarea.Value()
			message := "# User\n" + prompt
			m.messages = append(m.messages, message)
			renderedMessages, err := glamour.Render(
				lipgloss.
					NewStyle().
					Width(m.viewport.Width).
					Render(strings.Join(m.messages, "\n")),
				"dark",
			)
			if err != nil {
				m.err = err
				return m, nil
			}
			m.viewport.SetContent(renderedMessages)
			m.textarea.Reset()
			m.viewport.GotoBottom()

			recvMsg, err := agent.OpenStream(prompt)
			if err != nil {
				m.err = err
				return m, nil
			}

			return m, func() tea.Msg {
				return *recvMsg
			}
		}

	case errMsg:
		m.err = msg
		return m, nil
	}

	batch := tea.Batch(tiCmd, vpCmd)

	return m, batch
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
