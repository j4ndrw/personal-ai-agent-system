package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/j4ndrw/personal-ai-agent-system/client/internal/ui"
)

func main() {
	p := tea.NewProgram(ui.InitialModel())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
