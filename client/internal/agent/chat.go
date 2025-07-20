package agent

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"net/http"

	"github.com/j4ndrw/personal-ai-agent-system/client/internal/async"
	"github.com/j4ndrw/personal-ai-agent-system/client/internal/state"
)

type OnChunk func(chunk *AgentChunk)

type ReceiveStreamChunkMsg struct {
	AgentChunk *AgentChunk
	Response   *http.Response
	Time       time.Time
}
type ReceiveStreamChunkTickMsg ReceiveStreamChunkMsg

func OpenStream(
	prompt string,
) (ReceiveStreamChunkMsg, error) {
	agentChunk := AgentChunk{}

	bodyData, err := json.Marshal(map[string]string{"prompt": prompt})
	if err != nil {
		return ReceiveStreamChunkMsg{}, err
	}

	resp, err := http.Post(
		Endpoint,
		"application/json",
		bytes.NewBuffer(bodyData),
	)
	if err != nil {
		return ReceiveStreamChunkMsg{}, err
	}

	return ReceiveStreamChunkMsg{
		AgentChunk: &agentChunk,
		Response:   resp,
		Time:       time.Now(),
	}, nil
}

func ReadChunk(msg *ReceiveStreamChunkMsg, rcStateNode *state.ReadChunk) {
	buf := make([]byte, 1024)
	_, err := msg.Response.Body.Read(buf)

	if err == io.EOF {
		msg.Response.Body.Close()

		rcStateNode.Result = nil
		rcStateNode.Err = nil
		rcStateNode.Phase = async.DoneAsyncResultState
		return
	}
	if err != nil {
		rcStateNode.Result = nil
		rcStateNode.Err = err
		rcStateNode.Phase = async.DoneAsyncResultState
		return
	}

	buf = bytes.Trim(buf, "\x00")

	if len(buf) == 0 {
		rcStateNode.Result = nil
		rcStateNode.Err = nil
		rcStateNode.Phase = async.DoneAsyncResultState
		return
	}

	err = msg.AgentChunk.ParseAgentChunk(&buf)
	if err != nil {
		rcStateNode.Result = nil
		rcStateNode.Err = err
		rcStateNode.Phase = async.DoneAsyncResultState
		return
	}

	rcStateNode.Result = msg
	rcStateNode.Err = nil
	rcStateNode.Phase = async.DoneAsyncResultState
}

func MapAnswer(chunk *AgentChunk, thinking *bool) string {
	if *thinking != chunk.Answer.Thinking {
		*thinking = chunk.Answer.Thinking
	}
	return chunk.Answer.Content
}

func MapToolCall(chunk *AgentChunk) string {
	return "`" + chunk.ToolCall.ToolCall + " tool`\n```json\n" + chunk.ToolCall.JSONResult + "\n```\n\n"
}

func MapChunk(
	mappedChunk *string,
	toolCalls *[]string,
	thinking *bool,
) OnChunk {
	return func(chunk *AgentChunk) {
		if chunk.Answer == nil && chunk.ToolCall == nil {
			return
		}

		if chunk.Answer != nil && chunk.Answer.Content != "" {
			*mappedChunk = MapAnswer(chunk, thinking)
		} else if chunk.ToolCall != nil && chunk.ToolCall.ToolCall != "" {
			toolCall := MapToolCall(chunk)
			if len(*toolCalls) == 0 || (*toolCalls)[len(*toolCalls)-1] != toolCall {
				*toolCalls = append(*toolCalls, MapToolCall(chunk))
			}
		}
	}
}

func ProcessChunk(sink *[]string, chunk string, render func() error) error {
	if len(*sink) == 0 {
		*sink = append(*sink, chunk)
	} else {
		(*sink)[len(*sink)-1] = (*sink)[len(*sink)-1] + chunk
	}
	err := render()
	if err != nil {
		return err
	}
	return nil
}
