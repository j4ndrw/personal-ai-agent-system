package agent

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	// "log"

	// "io"
	"net/http"
	"strings"

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
	sb := ""
	if *thinking == chunk.Answer.Thinking {
		sb += chunk.Answer.Content
		return sb
	}

	*thinking = chunk.Answer.Thinking
	if *thinking {
		sb += "\n**Thinking**\n\n"
	} else {
		sb += "\n**Done thinking**\n\n"
	}
	sb += chunk.Answer.Content
	return sb
}

func MapToolCall(chunk *AgentChunk) string {
	return "`" + chunk.ToolCall.ToolCall + " tool`\n```json\n" + chunk.ToolCall.JSONResult + "\n```\n\n"
}

func MapChunk(mappedChunk *string, toolCalls *[]string, thinking *bool) OnChunk {
	return func(chunk *AgentChunk) {
		if chunk.Answer == nil && chunk.ToolCall == nil {
			return
		}

		if chunk.Answer != nil {
			*mappedChunk = MapAnswer(chunk, thinking)
		}

		if chunk.ToolCall != nil {
			*toolCalls = append(*toolCalls, MapToolCall(chunk))
		}
	}
}

func ProcessAnswerChunk(messages *[]string, chunk string, render func() error) error {
	(*messages)[len(*messages)-1] = (*messages)[len(*messages)-1] + chunk
	err := render()
	if err != nil {
		return err
	}
	return nil
}

func ProcessToolCalls(messages *[]string, toolCalls *[]string, render func() error) error {
	(*messages)[len(*messages)-1] = (*messages)[len(*messages)-1] + "\n" + strings.Join(*toolCalls, "\n")
	err := render()
	if err != nil {
		return err
	}
	return nil
}
