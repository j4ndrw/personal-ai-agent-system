package ui

import (
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/j4ndrw/personal-ai-agent-system/client/internal/agent"
	"golang.org/x/term"
)

type Model struct {
	viewport viewport.Model
	textarea textarea.Model
	spinner  spinner.Model
	state    State
}

func InitialModel() Model {
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		log.Fatal(err)
	}

	ta := TextAreaComponent("Chat with the agent system...", w, TextAreaHeight)
	sp := SpinnerComponent()
	vp := viewport.New(w, h-ta.Height()-lipgloss.Height(Gap))

	return Model{
		textarea: ta,
		spinner:  sp,
		viewport: vp,
		state: State{
			messages:          []string{},
			toolCalls:         []string{},
			agentMessageChunk: "",
			agentThinking:     false,
			err:               nil,
		},
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
		return m.ReceiveStreamChunkUpdate(msg)

	case tea.WindowSizeMsg:
		return m.WindowSizeUpdate(msg)

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m.QuitKeyUpdate()

		case tea.KeyEnter:
			return m.ChatMessageSendUpdate()
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case ErrMsg:
		m.state.err = msg
		return m, nil
	}

	batch := tea.Batch(tiCmd, vpCmd)
	return m, batch
}
