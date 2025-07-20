package state

type AgentMessageToShow int
const (
	AgentMessageShowThoughts AgentMessageToShow = iota
	AgentMessageShowAnswers
	AgentMessageShowToolCalls
)
