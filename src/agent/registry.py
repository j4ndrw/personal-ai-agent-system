from typing import Callable

from ollama import Message

from src.models.agent.answer import Answer

agent_registry: dict[
    str,
    Callable[[list[Message], Callable[[str], None]], tuple[Answer, str | None, bool]],
] = {}
