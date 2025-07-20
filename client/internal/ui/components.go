package ui

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/lipgloss"
)

func TextAreaComponent(placeholder string, width int, height int) textarea.Model {
	ta := textarea.New()
	ta.Placeholder = placeholder
	ta.Focus()

	ta.Prompt = PromptPrefix

	ta.SetWidth(width)
	ta.SetHeight(height)

	baseStyle := lipgloss.NewStyle()
	ta.FocusedStyle.CursorLine = baseStyle
	ta.BlurredStyle.Base = baseStyle.Faint(true)

	ta.ShowLineNumbers = true

	ta.KeyMap.InsertNewline.SetEnabled(true)

	return ta
}

func SpinnerComponent() spinner.Model {
	sp := spinner.New()
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	sp.Spinner = spinner.Points
	return sp
}
