package stringtransforms

import (
	"errors"
	"slices"
	"strings"

	"github.com/j4ndrw/personal-ai-agent-system/client/internal/agent"
)

func MapUserMessage(prompt string) string {
	sb := ""
	for _, line := range strings.Split(prompt, "\n") {
		sb += "> " + line + "\n"
	}
	return sb
}

func ExtractAgentAndPrompt(s string, agents *agent.Agents) (string, string, error) {
	tokens := strings.SplitN(s, " ", 2)
	errMsg := "Since you are in Agentic (manual) mode, you must specify an agent to target before the prompt. E.g. `@web_search who is grace hopper`"
	if len(tokens) == 1 {
		return "", "", errors.New(errMsg)
	}

	firstToken := tokens[0]
	secondToken := tokens[1]

	if rune(firstToken[0]) != '@' || len(firstToken) == 1 {
		return "", "", errors.New(errMsg)
	}

	agentToTarget := firstToken[1:]
	if !slices.Contains(*agents, agentToTarget) {
		return "", "", errors.New(errMsg)
	}


	if secondToken == "" {
		return "", "", errors.New(errMsg)
	}

	return agentToTarget, secondToken, nil
}
