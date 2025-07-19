package ui

import (
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/j4ndrw/personal-ai-agent-system/client/internal/agent"
	"github.com/j4ndrw/personal-ai-agent-system/client/internal/state"
	"golang.org/x/term"
)

type Model struct {
	viewport viewport.Model
	textarea textarea.Model
	help     help.Model
	spinner  spinner.Model
	state    state.State
}

func InitialModel() Model {
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		log.Fatal(err)
	}

	ta := TextAreaComponent("Chat with the agent system...", w, TextAreaHeight)
	sp := SpinnerComponent()
	vp := viewport.New(w, h-ta.Height()-1-lipgloss.Height(Gap))

	return Model{
		textarea: ta,
		spinner:  sp,
		viewport: vp,
		help:     help.New(),
		state: state.State{
			Messages: []string{},
			Agent: state.AgentState{
				ToolCalls: []string{},
				Token:     "",
				Thinking:  false,
			},
			Err:     nil,
			Waiting: false,
			Async:   state.AsyncState{},
		},
	}
}

func (m Model) Init() tea.Cmd {
	return textarea.Blink
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

	case agent.ReceiveStreamChunkTickMsg:
		return m.ReceiveStreamChunkTickUpdate(msg)

	case tea.WindowSizeMsg:
		return m.WindowSizeUpdate(msg)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Quit):
			return m.QuitKeyUpdate()

		case key.Matches(msg, Keys.SendMessage):
			return m.ChatMessageSendUpdate()

		case key.Matches(msg, Keys.ScrollUp):
			return m.ScrollUpUpdate()

		case key.Matches(msg, Keys.ScrollDown):
			return m.ScrollDownUpdate()
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case ErrMsg:
		m.state.Err = msg
		return m, nil
	}

	cmds := []tea.Cmd{tiCmd, vpCmd}
	if m.state.Waiting {
		cmds = append(cmds, m.spinner.Tick)
	}

	batch := tea.Batch(cmds...)
	return m, batch
}

func (m Model) View() string {
	return fmt.Sprintf(
		"%s%s%s\n  %s",
		m.viewport.View(),
		Gap,
		func() string {
			if !m.state.Waiting {
				return m.textarea.View()
			}

			spinnerText := map[bool]string{
				true:  "Thinking",
				false: "Generating",
			}
			return lipgloss.
				NewStyle().
				Faint(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Render(
					fmt.Sprintf(
						"%s%s %s",
						PromptPrefix,
						spinnerText[m.state.Agent.Thinking],
						m.spinner.View(),
					),
				)
		}(),
		m.help.View(Keys),
	)
}
