package agent

import (
	"encoding/json"
	"errors"
)

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
	answerErr := ac.ParseAnswer(body)
	toolCallErr := ac.ParseToolCall(body)

	if answerErr != nil && toolCallErr != nil {
		return errors.New("Could not parse agent response chunk")
	}
	return nil
}
