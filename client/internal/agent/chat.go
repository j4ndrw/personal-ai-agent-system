package agent

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
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

func MapChunk(mappedChunk *string, thinking *bool) OnChunk {
	return func(chunk AgentChunk) {
		if chunk.Answer == nil && chunk.ToolCall == nil {
			return
		}

		if *mappedChunk == "" {
			*mappedChunk = "# Agent\n"
		}

		if chunk.ToolCall != nil {
			*mappedChunk = "`tool:`" + chunk.ToolCall.ToolCall + "\n```json\n" + chunk.ToolCall.JSONResult + "```\n\n"
			return
		}

		if *thinking != chunk.Answer.Thinking == true {
			*thinking = chunk.Answer.Thinking
			if *thinking {
				*mappedChunk = "**Thinking**\n\n"
			} else {
				*mappedChunk = "**Done thinking**\n\n"
			}
		}

		*mappedChunk = chunk.Answer.Content
	}
}

func ProcessChunk(messages *[]string, chunk string, render func() error) error {
	(*messages)[len(*messages)-1] = (*messages)[len(*messages)-1]+chunk
	err := render()
	if err != nil {
		return err
	}
	return nil
}
