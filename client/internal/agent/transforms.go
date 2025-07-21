package agent

import "github.com/j4ndrw/personal-ai-agent-system/client/internal/state"

type AgentSink string
type SinkFactory func() (*[]string, string)
type SinkMap map[AgentSink]SinkFactory

const (
	AgentThoughtsSink AgentSink = "thoughts"
	AgentAnswersSink  AgentSink = "answers"
	AgentToolCallSink AgentSink = "toolcalls"
)

func MapAgentAnswer(chunk AgentChunk, thinking bool) (string, bool) {
	if thinking != chunk.Answer.Thinking {
		thinking = chunk.Answer.Thinking
	}
	return chunk.Answer.Content, thinking
}

func MapAgentToolCall(chunk AgentChunk) string {
	return "`" + chunk.ToolCall.ToolCall + " tool`\n```json\n" + chunk.ToolCall.JSONResult + "\n```\n\n"
}

func (sm *SinkMap) MapAgentChunk(
	chunk AgentChunk,
	state *state.AgentState,
) *SinkMap {
	state.ChunkId = chunk.Id
	state.Token = ""
	state.ToolCall = ""

	if chunk.Type == "answer" && chunk.Answer.Content != "" {
		state.Token, state.Thinking = MapAgentAnswer(chunk, state.Thinking)
		state.ToolCall = ""
	}

	if chunk.Type == "tool_call" && chunk.ToolCall.ToolCall != "" {
		state.ToolCall = MapAgentToolCall(chunk)
		state.Token = ""
		state.SelectedAgent = chunk.AgentName
	}

	return sm
}

func CreateSinkMap(st *state.State) *SinkMap {
	return &SinkMap{
		AgentThoughtsSink: func() (*[]string, string) {
			return &st.AgentThoughts, st.Agent.Token
		},
		AgentAnswersSink: func() (*[]string, string) {
			return &st.AgentAnswers, st.Agent.Token
		},
		AgentToolCallSink: func() (*[]string, string) {
			return &st.AgentToolCalls, st.Agent.ToolCall
		},
	}
}

func (sm *SinkMap) ProcessChunk(st state.State, process func(sink *[]string, chunk string) error) error {
	for sink, _ := range *sm {
		thoughts := sink == AgentThoughtsSink && st.Agent.Thinking && st.Agent.Token != ""
		answers := sink == AgentAnswersSink && !st.Agent.Thinking && st.Agent.Token != ""
		toolcalls := sink == AgentToolCallSink && st.Agent.ToolCall != ""

		if thoughts || answers || toolcalls {
			sink, chunk := (*sm)[sink]()
			err := process(sink, chunk)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
