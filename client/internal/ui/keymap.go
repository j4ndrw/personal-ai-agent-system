package ui

import "github.com/charmbracelet/bubbles/key"

type SharedKeyMap struct {
	ScrollUp   key.Binding
	ScrollDown key.Binding
	Quit       key.Binding
}
type NormalModeKeyMap struct {
	InspectThoughts  key.Binding
	InspectAnswers   key.Binding
	InspectToolCalls key.Binding
	Yank             key.Binding
	ToInsertMode     key.Binding
	ScrollToTop      key.Binding
	ScrollToBottom   key.Binding
	CycleChatModes   key.Binding
	SharedKeyMap
}
type InsertModeKeyMap struct {
	ToNormalMode key.Binding
	SendMessage  key.Binding
	SharedKeyMap
}

var SharedKeys = SharedKeyMap{
	ScrollUp:   key.NewBinding(key.WithKeys("ctrl+u"), key.WithHelp("ctrl+u", "Scroll up")),
	ScrollDown: key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("ctrl+d", "Scroll down")),
	Quit:       key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "Quit")),
}
var NormalModeKeys = NormalModeKeyMap{
	ToInsertMode:     key.NewBinding(key.WithKeys("i"), key.WithHelp("i", "Insert mode")),
	InspectThoughts:  key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "(In Agentic mode) Inspect thoughts")),
	InspectAnswers:   key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "(In Agentic mode) Inspect answers")),
	InspectToolCalls: key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "(In Agentic mode) Inspect tool calls")),
	Yank:             key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "Yank")),
	ScrollToTop:      key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "Scroll to top")),
	ScrollToBottom:   key.NewBinding(key.WithKeys("G"), key.WithHelp("G", "Scroll to bottom")),
	CycleChatModes:   key.NewBinding(key.WithKeys(" "), key.WithHelp("spc", "Cycle chat modes")),
	SharedKeyMap:     SharedKeys,
}
var InsertModeKeys = InsertModeKeyMap{
	ToNormalMode: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "Normal mode")),
	SendMessage:  key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "Send prompt")),
	SharedKeyMap: SharedKeys,
}

func (k NormalModeKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.ScrollUp, k.ScrollDown, k.Quit, k.ToInsertMode}
}
func (k InsertModeKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.ScrollUp, k.ScrollDown, k.Quit, k.ToNormalMode, k.SendMessage}
}

func (k NormalModeKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.ScrollUp, k.ScrollDown, k.Quit},
		{k.ScrollToTop, k.ScrollToBottom, k.CycleChatModes},
		{k.InspectThoughts, k.InspectAnswers, k.InspectToolCalls},
		{k.Yank, k.ToInsertMode},
	}
}

func (k InsertModeKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.ScrollUp, k.ScrollDown, k.Quit},
		{k.SendMessage, k.ToNormalMode},
	}
}
