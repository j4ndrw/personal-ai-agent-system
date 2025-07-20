package state

import (
	"github.com/j4ndrw/personal-ai-agent-system/client/internal/async"
)

type AgentState struct {
	Token     string
	Thinking  bool
	ToolCalls []string
}

type ReadChunk struct {
	Result any
	Err    error
	Phase  async.AsyncResultState
}

type AsyncState struct {
	ReadChunk *ReadChunk
}

type State struct {
	UserMessages       []string
	AgentThoughts      []string
	AgentAnswers       []string
	AgentToolCalls     []string
	AgentMessageToShow AgentMessageToShow
	Err                error
	Waiting            bool
	Agent              AgentState
	Async              AsyncState
}
