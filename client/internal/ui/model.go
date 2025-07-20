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
	keys             help.KeyMap
	spinner          spinner.Model
	state            state.State
	markdownRenderer glamour.TermRenderer
}

func InitialModel() Model {
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		log.Fatal(err)
	}

	ta := TextAreaComponent("Chat with the agent system...", w, 1)
	sp := SpinnerComponent()
	vph := h - ta.Height() - 5 - lipgloss.Height(Gap)
	vp := viewport.New(w, vph)
	vp.Style = lipgloss.
		NewStyle().
		PaddingTop(2).
		PaddingLeft(1).
		PaddingRight(1)
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
		keys:             InsertModeKeys,
		markdownRenderer: *markdownRenderer,
		state: state.State{
			Mode:               state.InsertMode,
			UserMessages:       []string{},
			AgentThoughts:      []string{},
			AgentAnswers:       []string{},
			AgentToolCalls:     []string{},
			AgentMessageToShow: state.AgentMessageShowAnswers,
			Agent: state.AgentState{
				ToolCall: "",
				Token:    "",
				Thinking: false,
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

	switch m.state.Mode {
	case state.NormalMode:
		m.keys = NormalModeKeys
		break
	case state.InsertMode:
		m.keys = InsertModeKeys
		break
	}

	switch msg := msg.(type) {
	case agent.ReceiveStreamChunkMsg:
		return m.ReceiveStreamChunkUpdate(msg)

	case agent.ReceiveStreamChunkTickMsg:
		return m.ReceiveStreamChunkTickUpdate(msg)

	case tea.WindowSizeMsg:
		return m.WindowSizeUpdate(msg)

	case tea.KeyMsg:
		switch keys := m.keys.(type) {
		case InsertModeKeyMap:
			switch {
			case key.Matches(msg, keys.Quit):
				return m.QuitKeyUpdate()

			case key.Matches(msg, keys.ScrollUp):
				return m.ScrollUpUpdate()

			case key.Matches(msg, keys.ScrollDown):
				return m.ScrollDownUpdate()

			case key.Matches(msg, keys.ToNormalMode):
				return m.ToNormalModeUpdate()

			case key.Matches(msg, keys.NewLine):
				return m.NewLineUpdate()
			}
		case NormalModeKeyMap:
			switch {
			case key.Matches(msg, keys.Quit):
				return m.QuitKeyUpdate()

			case key.Matches(msg, keys.ScrollUp):
				return m.ScrollUpUpdate()

			case key.Matches(msg, keys.ScrollDown):
				return m.ScrollDownUpdate()

			case key.Matches(msg, keys.SendMessage):
				return m.ChatMessageSendUpdate()

			case key.Matches(msg, keys.Yank):
				return m.YankUpdate()

			case key.Matches(msg, keys.InspectThoughts):
				return m.InspectThoughtsUpdate()

			case key.Matches(msg, keys.InspectAnswers):
				return m.InspectAnswersUpdate()

			case key.Matches(msg, keys.InspectToolCalls):
				return m.InspectToolCallsUpdate()

			case key.Matches(msg, keys.ToInsertMode):
				return m.ToInsertModeUpdate()

			case key.Matches(msg, keys.ScrollToTop):
				return m.ScrollToTopUpdate()

			case key.Matches(msg, keys.ScrollToBottom):
				return m.ScrollToBottomUpdate()
			}
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case ErrMsg:
		m.state.Err = msg
		return m, nil
	}

	cmds := []tea.Cmd{tiCmd, vpCmd, m.spinner.Tick}
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
						"%s  %s %s",
						PromptPrefix,
						spinnerText[m.state.Agent.Thinking],
						m.spinner.View(),
					),
				)
		}(),
		lipgloss.
			NewStyle().
			PaddingLeft(4).
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
			PaddingLeft(4).
			Render(
				strings.Join(strings.Split(m.help.View(m.keys), "\r\n"), "   \n"),
			),
	)
}
