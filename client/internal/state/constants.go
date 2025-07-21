package state

type AgentMessageToShow int

const (
	AgentMessageShowThoughts AgentMessageToShow = iota
	AgentMessageShowAnswers
	AgentMessageShowToolCalls
)

type Mode int

const (
	NormalMode Mode = iota
	InsertMode
)

type ChatMode int

const (
	SimpleChatMode  ChatMode = iota
	AgenticAutoChatMode
	AgenticManualChatMode
)
