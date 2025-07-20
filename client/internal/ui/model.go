package ui

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/j4ndrw/personal-ai-agent-system/client/internal/agent"
	"github.com/j4ndrw/personal-ai-agent-system/client/internal/state"
	"golang.org/x/term"
)

type Model struct {
	viewport         viewport.Model
	textarea         textarea.Model
	help             help.Model
	spinner          spinner.Model
	state            state.State
	markdownRenderer glamour.TermRenderer
}

func InitialModel() Model {
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		log.Fatal(err)
	}

	ta := TextAreaComponent("Chat with the agent system...", w, TextAreaHeight)
	sp := SpinnerComponent()
	vp := viewport.New(w, h-ta.Height()-3-lipgloss.Height(Gap))
	vp.Style = lipgloss.NewStyle().PaddingTop(2).PaddingLeft(1).PaddingRight(1).Width(w).Align(lipgloss.Center)
	hlp := help.New()
	hlp.ShowAll = true

	markdownRenderer, err := glamour.NewTermRenderer(glamour.WithAutoStyle(), glamour.WithWordWrap(w))
	if err != nil {
		log.Fatal(err)
	}

	return Model{
		textarea:         ta,
		spinner:          sp,
		viewport:         vp,
		help:             hlp,
		markdownRenderer: *markdownRenderer,
		state: state.State{
			UserMessages:       []string{},
			AgentThoughts:      []string{},
			AgentAnswers:       []string{},
			AgentToolCalls:     []string{},
			AgentMessageToShow: state.AgentMessageShowAnswers,
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

		case key.Matches(msg, Keys.Yank):
			return m.YankUpdate()

		case key.Matches(msg, Keys.InspectThoughts):
			return m.InspectThoughtsUpdate()

		case key.Matches(msg, Keys.InspectAnswers):
			return m.InspectAnswersUpdate()

		case key.Matches(msg, Keys.InspectToolCalls):
			return m.InspectToolCallsUpdate()
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
		"%s%s%s\n%s\n%s",
		m.viewport.View(),
		Gap,
		func() string {
			if !m.state.Waiting {
				return lipgloss.
					NewStyle().
					Render(m.textarea.View())
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
		lipgloss.
			NewStyle().
			PaddingLeft(3).
			Render(
				func() string {
					inspecting := fmt.Sprintf(
						"Currently inspecting: Agent %s",
						func() string {
							switch m.state.AgentMessageToShow {
							case state.AgentMessageShowAnswers:
								return "answers"
							case state.AgentMessageShowThoughts:
								return "thoughts"
							default:
								return "tool calls"
							}
						}(),
					)
					return lipgloss.
						NewStyle().
						Faint(true).
						Foreground(lipgloss.Color("#FFFFFF")).
						Render(
							fmt.Sprintf(
								"%s",
								strings.Join(
									[]string{
										inspecting,
									},
									" | ",
								),
							),
						)
				}(),
			),
		lipgloss.
			NewStyle().
			PaddingLeft(3).
			Render(
				strings.Join(strings.Split(m.help.View(Keys), "\r\n"), "   \n"),
			),
	)
}
