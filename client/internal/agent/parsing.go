package agent

import (
	"encoding/json"
)

type AgentChunk struct {
	Type string `json:"type"`
	Answer
	ToolCall
}

func (ac *AgentChunk) ParseAgentChunk(body *[]byte) error {
	err := json.Unmarshal(*body, &ac)
	if err != nil {
		return err
	}
	return nil
}
