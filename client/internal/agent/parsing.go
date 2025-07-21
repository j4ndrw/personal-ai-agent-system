package agent

import (
	"encoding/json"
)

type AgentChunk struct {
	Type string `json:"type"`
	Id   string `json:"id"`
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

func (a *Agents) ParseAgents(body *[]byte) error {
	err := json.Unmarshal(*body, &a)
	if err != nil {
		return err
	}
	return nil
}
