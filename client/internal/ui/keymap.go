package ui

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	ScrollUp    key.Binding
	ScrollDown  key.Binding
	Quit        key.Binding
	SendMessage key.Binding
}

var Keys = KeyMap{
	ScrollUp:    key.NewBinding(key.WithKeys("ctrl+u"), key.WithHelp("ctrl+u", "Scroll Up")),
	ScrollDown:  key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("ctrl+d", "Scroll Down")),
	Quit:        key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "Quit")),
	SendMessage: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "Send prompt")),
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.ScrollUp, k.ScrollDown, k.SendMessage, k.Quit}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}
