package state

import (
	"github.com/j4ndrw/personal-ai-agent-system/client/internal/async"
)

type AgentState struct {
	ProcessedChunkIds []string
	ChunkId           string
	Token             string
	Thinking          bool
	ToolCall          string
	SelectedAgent         string
}

type ReadChunkData struct {
	Result any
	Err    error
	Phase  async.AsyncResultState
}

type ReadChunk struct {
	Data *ReadChunkData
}

type AsyncState struct {
	ReadChunk ReadChunk
}

type State struct {
	UserMessages       []string
	AgentThoughts      []string
	AgentAnswers       []string
	AgentToolCalls     []string
	AgentMessageToShow AgentMessageToShow
	ChatMode           ChatMode
	Agents             []string
	Err                error
	Waiting            bool
	Agent              AgentState
	Async              AsyncState
	Mode               Mode
}
