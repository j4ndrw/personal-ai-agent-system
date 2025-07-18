package agent

type Answer struct {
	Type string `json:"type" validate:"oneof=answer,required"`
    Thinking bool `json:"thinking" validate:"required"`
    Content string `json:"content" validate:"required"`
}

type ToolCall struct {
	Type string `json:"type" validate:"oneof=tool_call,required"`
    ToolCall string `json:"tool_call" validate:"required"`
    JSONResult string `json:"result" validate:"required"`
}
