import json
from typing import Any, Generator
import uuid

from ollama import Message
from src.client import ollama_client
from src.tools.tools import ToolRepository
from src.utils import StatefulGenerator


def create_chat_handler(
    *,
    tool_repository: ToolRepository | None = None
):
    def _chat(
        history: list[Message],
        model: str,
        with_tools: bool,
        think: bool,
        mark_as_thinking: bool = False
    ) -> Generator[str, Any, Message]:
        final_message: Message = Message(role="assistant")
        stream = ollama_client.chat(
            model=model,
            messages=history,
            think=think,
            stream=True,
            tools=None if not with_tools or tool_repository is None else [*tool_repository.values()],
        )
        for chunk in stream:
            if chunk.message.content:
                yield f"{json.dumps({'id': str(uuid.uuid4()), 'type': 'answer', 'thinking': True if mark_as_thinking else think, 'content': chunk.message.content})}\n"
                final_message.content = (
                    final_message.content + chunk.message.content
                    if final_message.content
                    else chunk.message.content
                )
            if chunk.message.tool_calls:
                for tool_call in chunk.message.tool_calls:
                    final_message.tool_calls = (
                        [tool_call]
                        if final_message.tool_calls is None
                        else [*final_message.tool_calls, tool_call]
                    )
        return final_message

    def chat(
        *,
        history: list[Message],
        model: str,
        with_tools: bool,
        think: bool,
        mark_as_thinking: bool = False
    ):
        return StatefulGenerator(
            _chat(history, model, with_tools, think, mark_as_thinking)
        )  # pyright: ignore

    return chat
