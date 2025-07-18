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
	var answer Answer
	err := json.Unmarshal(*body, &answer)
	if err != nil {
		return err
	}
	ac.Answer = &answer
	return nil
}

func (ac *AgentChunk) ParseToolCall(body *[]byte) error {
	var toolCall ToolCall
	err := json.Unmarshal(*body, &toolCall)
	if err != nil {
		return err
	}
	ac.ToolCall = &toolCall
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
