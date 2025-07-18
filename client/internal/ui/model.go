package ui

import (
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/j4ndrw/personal-ai-agent-system/client/internal/agent"
	"golang.org/x/term"
)

type Model struct {
	viewport          viewport.Model
	textarea          textarea.Model
	messages          []string
	toolCalls         []string
	agentMessageChunk string
	agentThinking     bool
	err               error
}

func InitialModel() Model {
	ta := textarea.New()
	ta.Placeholder = "Chat with the agent..."
	ta.Focus()

	ta.Prompt = "┃ "

	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		log.Fatal(err)
	}
	ta.SetWidth(w)
	ta.SetHeight(TextareaHeight)
	vp := viewport.New(w, h-ta.Height()-lipgloss.Height(Gap))
	ta.KeyMap.InsertNewline.SetEnabled(false)

	return Model{
		textarea:          ta,
		messages:          []string{},
		toolCalls:         []string{},
		agentMessageChunk: "",
		agentThinking:     false,
		viewport:          vp,
		err:               nil,
	}
}

func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

func (m Model) View() string {
	return fmt.Sprintf(
		"%s%s%s",
		m.viewport.View(),
		Gap,
		m.textarea.View(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case agent.ReceiveStreamChunkMsg:
		cmd, err := m.ReceiveStreamChunkHandler(msg)
		if err != nil {
			m.err = err
			return m, nil
		}
		if cmd != nil {
			return m, cmd
		}

	case tea.WindowSizeMsg:
		err := m.WindowSizeHandler(msg)
		if err != nil {
			m.err = err
			return m, nil
		}

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.QuitKeyHandler()
			return m, tea.Quit

		case tea.KeyEnter:
			cmd, err := m.ChatMessageSendHandler()
			if err != nil {
				m.err = err
				return m, nil
			}
			return m, cmd
		}

	case ErrMsg:
		m.err = msg
		return m, nil
	}

	batch := tea.Batch(tiCmd, vpCmd)

	return m, batch
}
