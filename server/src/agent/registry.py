from typing import Any, Callable, Generator

from ollama import Message

from src.models.agent.answer import Answer

agent_registry: dict[
    str,
    Callable[
        [list[Message]],
        Generator[str, Any, tuple[Answer, str | None, bool]],
    ],
] = {}
