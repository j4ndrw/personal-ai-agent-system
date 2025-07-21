package agent

type Answer struct {
	Thinking bool   `json:"thinking"`
	Content  string `json:"content"`
	AgentName string `json:"agent_name"`
}

type ToolCall struct {
	ToolCall   string `json:"tool_call"`
	JSONResult string `json:"result"`
}

type Agents []string
