import json
import os
from typing import Callable

from ollama import Message

from src.tools.tools import ToolHandlers, ToolRepository, load_toolkits
from src.utils import load_model
from src.client import ollama_client

agents: dict[str, Callable[[list[Message]], list[Message]]] = {}


def register_agent(
    *,
    name: str,
    when_to_dispatch: str = "",
    model: str,
    toolkits: list[str] | None = None,
):
    loaded_toolkits = (
        load_toolkits(
            os.path.join(".", "src", "tools", "toolkits"),
            list(map(lambda toolkit: f"{toolkit}.py", toolkits)),  # pyright: ignore
        )
        if toolkits is None or len(toolkits) == 0
        else []
    )
    load_model(model=model)

    def run_agent(history: list[Message]) -> list[Message]:
        history_snapshot = [*history]

        tool_repository: ToolRepository = {}
        tool_handlers: ToolHandlers = {}
        for tool in loaded_toolkits:
            tool_repository = {**tool_repository, **tool.repository}
            tool_handlers = {**tool_handlers, **tool.handlers}

        chat = lambda tool_repository: ollama_client.chat(
            model=model,
            messages=history_snapshot,
            tools=None if tool_repository is None else [*tool_repository.values()],
        )
        message = chat(tool_repository).message
        tool_calls = [
            *filter(
                lambda tool_call: tool_call.function.name in tool_repository,
                message.tool_calls or [],
            )
        ]

        new_history = [message]

        if len(tool_calls) == 0:
            new_history = [*new_history, chat(None).message]
            return new_history

        found_tool = False
        for tool_call in tool_calls:
            function_to_call: Callable | None = tool_repository.get(
                tool_call.function.name
            )  # pyright: ignore
            if function_to_call is None:
                continue
            for [tool, args] in tool_handlers.items():
                if function_to_call.__name__ == tool.__name__:
                    found_tool = True
                    result = tool(*args(tool_call))
                    tool_message = Message(
                        role="tool", content=json.dumps(result), tool_calls=[tool_call]
                    )
                    new_history = [*new_history, tool_message]
                    break

        if found_tool:
            new_history = [
                *new_history,
                ollama_client.chat(
                    model=model,
                    messages=[*history_snapshot, *new_history],
                ).message,
            ]
        return new_history

    run_agent.name = name  # pyright: ignore
    run_agent.when_to_dispatch = when_to_dispatch  # pyright: ignore
    agents[name] = run_agent
    return run_agent
