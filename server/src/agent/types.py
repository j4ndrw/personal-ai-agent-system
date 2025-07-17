from typing import Any, Callable, Generator

from ollama import Message

DispatchedAgent = str | None
ToolCallResults = list[tuple[str, str]]
AgentStream = Generator[str, Any, tuple[DispatchedAgent, bool, ToolCallResults]]
Agent = Callable[[list[Message]], AgentStream]
AgentRegistry = dict[str, Agent]
