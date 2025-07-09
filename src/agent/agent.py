import json
import os
from typing import Callable

from ollama import Message

from src.client import ollama_client
from src.constants import INTERPRETATION_MODEL, NON_AGENTIC_MODEL, ROUTER_MODEL
from src.models.agent.answer import Answer
from src.models.chat.options import Options
from src.tools.tools import ToolHandlers, ToolRepository, load_toolkits
from src.utils import load_model

agent_registry: dict[
    str, Callable[[list[Message], Options | None], tuple[Answer, str | None]]
] = {}


def register_agent(
    *,
    name: str,
    when_to_dispatch: str = "",
    toolkits: list[str] | None = None,
):
    loaded_toolkits = (
        load_toolkits(
            os.path.join(".", "src", "tools", "toolkits"),
            list(
                map(lambda toolkit: f"{toolkit}.py", toolkits or [])
            ),  # pyright: ignore
        )
        if toolkits is not None and len(toolkits) > 0
        else []
    )

    for model in [ROUTER_MODEL, INTERPRETATION_MODEL, NON_AGENTIC_MODEL]:
        load_model(model=model)

    tool_repository: ToolRepository = {}
    tool_handlers: ToolHandlers = {}
    for toolkit in loaded_toolkits:
        tool_repository = {**tool_repository, **toolkit.repository}
        tool_handlers = {**tool_handlers, **toolkit.handlers}

    def chat(
        *,
        history: list[Message],
        model: str,
        with_tools: bool,
        think: bool,
        options: Options | None = None,
    ):
        _options = (
            {
                "temperature": options.temperature,
                "max_tokens": options.max_tokens,
                "n": options.n,
                "presence_penalty": options.presence_penalty,
                "frequency_penalty": options.frequency_penalty,
                "top_p": options.top_p,
            }
            if options is not None
            else None
        )
        return ollama_client.chat(
            model=model,
            messages=history,
            think=think,
            stream=False,
            tools=None if not with_tools else [*tool_repository.values()],
            options=_options,
        )

    def run_agent(
        history: list[Message], options: Options | None = None
    ) -> tuple[Answer, str | None]:
        print(f"Running `{name}` agent...")
        answer = Answer()

        def maybe_agentic_response():
            print(
                "\t[ROUTING] Determining whether the response should be agentic or not..."
            )
            message = chat(
                history=history,
                model=ROUTER_MODEL,
                with_tools=True,
                think=True,
                options=options,
            ).message

            print(
                f"\t[TOOL CALL] Detected tool calls: {[tool_call.function.name for tool_call in (message.tool_calls or [])]}"
            )
            tool_calls = [
                *filter(
                    lambda tool_call: tool_call.function.name in tool_repository,
                    message.tool_calls or [],
                )
            ]
            is_agentic = len(tool_calls) > 0
            if is_agentic:
                answer.agentic_message = message

            return tool_calls, is_agentic

        def non_agentic_response():
            print("\t[NON-AGENTIC] Appending non-agentic response...")
            answer.non_agentic_message = chat(
                history=history,
                model=NON_AGENTIC_MODEL,
                with_tools=False,
                think=False,
                options=options,
            ).message

        def try_to_execute_tool_calls(tool_calls: list[Message.ToolCall]):
            success = False
            dispatched_agent = None

            for tool_call in tool_calls:
                function_to_call: Callable | None = tool_repository.get(
                    tool_call.function.name
                )  # pyright: ignore
                if function_to_call is None:
                    continue
                for [tool, args] in tool_handlers.items():
                    if function_to_call.__name__ == tool.__name__:
                        print(
                            "\t[AGENTIC] Tool found - appending result of tool call to history..."
                        )
                        success = True
                        result = tool(*args(tool_call))
                        answer.tool_result_message[tool.__name__] = Message(
                            role="tool",
                            content=json.dumps(result, indent=4),
                            tool_calls=[tool_call],
                        )

                        if tool.__name__ == "dispatch_agent":
                            dispatched_agent = result["agent_to_dispatch"]
                        break

            return success, dispatched_agent

        def interpret_tool_call_result():
            print("\t[NON-AGENTIC] Appending interpretation of tool to history...")
            answer.interpretation_message = chat(
                history=[
                    *history,
                    *[
                        message
                        for message in [
                            answer.agentic_message,
                            answer.non_agentic_message,
                            *[
                                message
                                for message in answer.tool_result_message.values()
                            ],
                        ]
                        if message is not None
                    ],
                ],
                model=INTERPRETATION_MODEL,
                with_tools=False,
                think=False,
                options=options,
            ).message

        def force_dispatch_to_other_agent(agent: str):
            print(f"\t[DISPATCH] Forcing dispatch to `{agent}` agent...")
            answer.dispatch_message = Message(
                role="assistant",
                content=f"I need to delegate the task to the `{agent}` agent and determine what tool call to use, if applicable, and proceed from there...",
            )

        def pipeline():
            tool_calls, is_agentic = maybe_agentic_response()

            if not is_agentic:
                non_agentic_response()
                return answer, None

            success, dispatched_agent = try_to_execute_tool_calls(tool_calls)
            if success:
                if dispatched_agent:
                    force_dispatch_to_other_agent(dispatched_agent)
                else:
                    interpret_tool_call_result()

            return answer, dispatched_agent

        return pipeline()

    run_agent.name = name  # pyright: ignore
    run_agent.when_to_dispatch = when_to_dispatch  # pyright: ignore
    run_agent.available_tools = {  # pyright: ignore
        tool_name: tool_repository[tool_name].__doc__
        for tool_name in tool_repository.keys()
    }

    agent_registry[name] = run_agent
    return run_agent
