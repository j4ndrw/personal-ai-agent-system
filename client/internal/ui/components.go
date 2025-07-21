package ui

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

func TextInputComponent(placeholder string, width int) textinput.Model {
	ti := textinput.New()
	ti.SetValue("@")
	ti.Placeholder = placeholder
	ti.Focus()

	ti.Prompt = PromptPrefix
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))

	ti.ShowSuggestions = true
	ti.CharLimit = 8000

	ti.Width = width

	ti.Cursor.Style = lipgloss.NewStyle()
	ti.Cursor.TextStyle = lipgloss.NewStyle().Faint(true)

	return ti
}

func SpinnerComponent() spinner.Model {
	sp := spinner.New()
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	sp.Spinner = spinner.Points
	return sp
}
