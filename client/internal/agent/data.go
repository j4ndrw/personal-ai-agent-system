package agent

type Answer struct {
	Type string `json:"type"`
    Thinking bool `json:"thinking"`
    Content string `json:"content"`
}

type ToolCall struct {
	Type string `json:"type"`
    ToolCall string `json:"tool_call"`
    JSONResult string `json:"result"`
}
