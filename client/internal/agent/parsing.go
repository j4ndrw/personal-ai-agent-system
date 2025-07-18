package agent

import "encoding/json"

type AgentChunk struct {
	Answer   *Answer
	ToolCall *ToolCall
}

func (ac *AgentChunk) ParseAnswer(body *[]byte) error {
	err := json.Unmarshal(*body, &ac.Answer)
	if err != nil {
		return err
	}
	return nil
}

func (ac *AgentChunk) ParseToolCall(body *[]byte) error {
	err := json.Unmarshal(*body, &ac.ToolCall)
	if err != nil {
		return err
	}
	return nil
}

func (ac *AgentChunk) ParseAgentChunk(body *[]byte) error {
	err := ac.ParseAnswer(body)
	if err == nil {
		return nil
	}

	err = ac.ParseToolCall(body)
	if err == nil {
		return nil
	}

	return err
}
