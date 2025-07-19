package agent

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

type OnChunk func(chunk *AgentChunk)

type ReceiveStreamChunkMsg struct {
	AgentChunk *AgentChunk
	Response   *http.Response
}

func OpenStream(
	prompt string,
) (*ReceiveStreamChunkMsg, error) {
	agentChunk := AgentChunk{}

	bodyData, err := json.Marshal(map[string]string{"prompt": prompt})
	if err != nil {
		return &ReceiveStreamChunkMsg{}, err
	}

	resp, err := http.Post(
		Endpoint,
		"application/json",
		bytes.NewBuffer(bodyData),
	)
	if err != nil {
		return &ReceiveStreamChunkMsg{}, err
	}

	return &ReceiveStreamChunkMsg{
		AgentChunk: &agentChunk,
		Response:   resp,
	}, nil
}

func ReadChunk(msg ReceiveStreamChunkMsg, onChunk OnChunk) (*ReceiveStreamChunkMsg, error) {
	buf := make([]byte, 1024)
	_, err := msg.Response.Body.Read(buf)
	buf = bytes.Trim(buf, "\x00")

	if err == io.EOF {
		msg.Response.Body.Close()
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	err = msg.AgentChunk.ParseAgentChunk(&buf)
	if err != nil {
		return nil, err
	}
	onChunk(msg.AgentChunk)
	return &msg, nil
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
