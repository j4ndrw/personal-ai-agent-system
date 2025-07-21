package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/j4ndrw/personal-ai-agent-system/client/internal/state"
)

const LAYOUT = "%s%s%s\n%s\n%s"

func (m *Model) PromptView() string {
	if !m.state.Waiting {
		return lipgloss.
			NewStyle().
			Render(m.textinput.View())
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

}

func (m *Model) HelpView() string {
	return lipgloss.
		NewStyle().
		PaddingLeft(4).
		Render(
			strings.Join(
				strings.Split(m.help.View(m.keys),
					"\r\n",
				),
				"   \n",
			),
		)
}

func (m *Model) StatesView() string {
	return lipgloss.
		NewStyle().
		PaddingLeft(4).
		Render(
			func() string {
				chatMode := map[state.ChatMode]string{
					state.SimpleChatMode:        "Simple",
					state.AgenticAutoChatMode:   "Agentic (auto)",
					state.AgenticManualChatMode: "Agentic (manual)",
				}
				currentChatMode := fmt.Sprintf("Chat mode: %s", chatMode[m.state.ChatMode])
				selectedAgent := fmt.Sprintf("Active agent: %s", m.state.Agent.SelectedAgent)
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

				var view string
				if m.state.ChatMode == state.SimpleChatMode {
					view = fmt.Sprintf("%s", strings.Join([]string{currentChatMode}, " | "))
				} else if m.state.ChatMode == state.AgenticManualChatMode && m.state.Waiting && m.state.Agent.SelectedAgent != "" {
					view = fmt.Sprintf("%s", strings.Join([]string{currentChatMode, inspecting, selectedAgent}, " | "))
				} else {
					view = fmt.Sprintf("%s", strings.Join([]string{currentChatMode, inspecting}, " | "))
				}
				return lipgloss.
					NewStyle().
					Faint(true).
					Foreground(lipgloss.Color("#FFFFFF")).
					Render(view)
			}(),
		)
}
