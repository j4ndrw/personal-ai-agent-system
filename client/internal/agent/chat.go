package agent

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

type OnChunk func(chunk AgentChunk)

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
		"http://localhost:6969/api/chat",
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
	onChunk(*msg.AgentChunk)
	return &msg, nil
}

func MapChunk(mappedChunk *string, toolCalls *[]string, thinking *bool) OnChunk {
	return func(chunk AgentChunk) {
		if chunk.Answer == nil && chunk.ToolCall == nil {
			return
		}
		sb := ""

		if chunk.Answer != nil {
			if *thinking != chunk.Answer.Thinking {
				*thinking = chunk.Answer.Thinking
				if *thinking {
					sb += "\n**Thinking**\n\n"
				} else {
					sb += "\n**Done thinking**\n\n"
				}
			}
			sb += chunk.Answer.Content
		}

		if chunk.ToolCall != nil {
			*toolCalls = append(*toolCalls, "`"+chunk.ToolCall.ToolCall+" tool`\n```json\n"+chunk.ToolCall.JSONResult+"\n```\n\n")
		}

		*mappedChunk = sb
	}
}

func ProcessChunk(messages *[]string, toolCalls *[]string, chunk string, render func() error) error {
	sb := ""
	if len(*toolCalls) > 0 {
		sb = (*messages)[len(*messages)-1] + chunk + "\n" + strings.Join(*toolCalls, "\n")
	} else {
		sb = (*messages)[len(*messages)-1] + chunk
	}
	(*messages)[len(*messages)-1] = sb
	err := render()
	if err != nil {
		return err
	}
	return nil
}
