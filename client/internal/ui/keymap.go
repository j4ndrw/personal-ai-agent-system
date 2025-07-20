package ui

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	ScrollUp         key.Binding
	ScrollDown       key.Binding
	Quit             key.Binding
	SendMessage      key.Binding
	Yank             key.Binding
	InspectThoughts  key.Binding
	InspectAnswers   key.Binding
	InspectToolCalls key.Binding
}

var Keys = KeyMap{
	ScrollUp:         key.NewBinding(key.WithKeys("ctrl+u"), key.WithHelp("ctrl+u", "Scroll up")),
	ScrollDown:       key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("ctrl+d", "Scroll down")),
	Quit:             key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "Quit")),
	SendMessage:      key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "Send prompt")),
	Yank:             key.NewBinding(key.WithKeys("ctrl+y"), key.WithHelp("ctrl+y", "Yank")),
	InspectThoughts:  key.NewBinding(key.WithKeys("ctrl+t"), key.WithHelp("ctrl+t", "Inspect thoughts")),
	InspectAnswers:   key.NewBinding(key.WithKeys("ctrl+a"), key.WithHelp("ctrl+a", "Inspect answers")),
	InspectToolCalls: key.NewBinding(key.WithKeys("ctrl+f"), key.WithHelp("ctrl+f", "Inspect tool calls")),
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.ScrollUp, k.ScrollDown, k.Yank, k.InspectThoughts, k.InspectAnswers, k.InspectToolCalls}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.ScrollUp, k.ScrollDown, k.Yank},
		{k.InspectThoughts, k.InspectAnswers, k.InspectToolCalls},
	}
}
