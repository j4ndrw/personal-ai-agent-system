package agent

type Answer struct {
	Thinking bool   `json:"thinking"`
	Content  string `json:"content"`
}

type ToolCall struct {
	ToolCall   string `json:"tool_call"`
	JSONResult string `json:"result"`
}
