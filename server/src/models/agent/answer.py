from dataclasses import dataclass, field

import ollama


@dataclass
class Answer:
    agentic_message: ollama.Message | None = field(default=None)
    non_agentic_message: ollama.Message | None = field(default=None)
    tool_result_message: dict[str, ollama.Message] = field(default_factory=lambda: {})
    interpretation_message: ollama.Message | None = field(default=None)
    dispatch_message: ollama.Message | None = field(default=None)
