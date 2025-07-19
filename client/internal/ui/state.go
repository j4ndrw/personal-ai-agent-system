package ui

type State struct {
	messages          []string
	toolCalls         []string
	agentMessageChunk string
	agentThinking     bool
	err               error
	waiting           bool
}
