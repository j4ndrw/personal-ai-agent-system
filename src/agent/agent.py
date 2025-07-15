import json
import os
from typing import Callable

from ollama import Message

from src.client import ollama_client
from src.constants import INTERPRETATION_MODEL, NON_AGENTIC_MODEL, ROUTER_MODEL
from src.models.agent.answer import Answer
from src.tools.tools import ToolHandlers, ToolRepository, load_toolkits
from src.utils import load_model
from src.tools.toolkits.router import dispatch_agent, mark_task_as_done
from src.agent.registry import agent_registry


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
        on_token: Callable[[str], None],
    ) -> Message:
        final_message: Message = Message(role="assistant")
        stream = ollama_client.chat(
            model=model,
            messages=history,
            think=think,
            stream=True,
            tools=None if not with_tools else [*tool_repository.values()],
        )
        for chunk in stream:
            if chunk.message.content:
                on_token(chunk.message.content)
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

    def run_agent(
        history: list[Message], on_token: Callable[[str], None]
    ) -> tuple[Answer, str | None, bool]:
        answer = Answer()

        def maybe_agentic_response():
            message = chat(
                history=history,
                model=ROUTER_MODEL,
                with_tools=True,
                think=True,
                on_token=on_token,
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
            is_task_done = True
            answer.non_agentic_message = chat(
                history=history,
                model=NON_AGENTIC_MODEL,
                with_tools=False,
                think=False,
                on_token=on_token,
            )
            return is_task_done

        def try_to_execute_tool_calls(tool_calls: list[Message.ToolCall]):
            success = False
            dispatched_agent = None
            is_task_done = False

            for tool_call in tool_calls:
                function_to_call: Callable | None = tool_repository.get(
                    tool_call.function.name
                )  # pyright: ignore
                if function_to_call is None:
                    continue
                for [tool, args] in tool_handlers.items():
                    if function_to_call.__name__ == tool.__name__:
                        success = True
                        result = tool(*args(tool_call))
                        answer.tool_result_message[tool.__name__] = Message(
                            role="tool",
                            content=json.dumps(result, indent=4),
                            tool_calls=[tool_call],
                        )

                        if tool.__name__ == dispatch_agent.__name__:
                            dispatched_agent = result["agent_to_dispatch"]
                        if tool.__name__ == mark_task_as_done.__name__:
                            is_task_done = result
                        break

            return success, dispatched_agent, is_task_done

        def interpret_tool_call_result():
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
                on_token=on_token,
            )

        def force_dispatch_to_other_agent(agent: str):
            answer.dispatch_message = Message(
                role="assistant",
                content=f"I need to delegate the task to the `{agent}` agent and determine what tool call to use, if applicable, and proceed from there...",
            )

        def pipeline():
            tool_calls, is_agentic = maybe_agentic_response()

            if not is_agentic:
                is_task_done = non_agentic_response()
                return answer, None, is_task_done

            success, dispatched_agent, is_task_done = try_to_execute_tool_calls(
                tool_calls
            )
            if success:
                if dispatched_agent:
                    force_dispatch_to_other_agent(dispatched_agent)
                else:
                    interpret_tool_call_result()

            return answer, dispatched_agent, is_task_done

        return pipeline()

    def check_if_task_is_done(history: list[Message]):
        content = f"""
Before proceeding I need to determine if the task is done. If it is done, I will call the {mark_task_as_done.__name__} function."
Definitions of done include:
- The user's question was answered
- The task was completed successfully
- There are no other paths to solve the problem at hand
        """
        is_done_message = Message(
            role="assistant",
            content=content,
        )
        history.append(is_done_message)

    run_agent.name = name  # pyright: ignore
    run_agent.when_to_dispatch = when_to_dispatch  # pyright: ignore
    run_agent.available_tools = {  # pyright: ignore
        tool_name: tool_repository[tool_name].__doc__
        for tool_name in tool_repository.keys()
    }
    run_agent.check_if_task_is_done = check_if_task_is_done  # pyright: ignore

    agent_registry[name] = run_agent
    return run_agent


def agentic_loop(
    history: list[Message],
    *,
    start_from_agent: Callable[
        [list[Message], Callable[[str], None]], tuple[Answer, str | None, bool]
    ],
    on_token: Callable[[str], None],
    max_loops: int = 10,
):
    answers: list[Answer] = []
    agent = start_from_agent
    epoch = 1
    while True:
        answer, dispatched_agent, is_task_done = agent(history, on_token)
        answers.append(answer)
        ai_messages = [
            message
            for message in [
                answer.agentic_message,
                answer.non_agentic_message,
                *[message for message in answer.tool_result_message.values()],
                answer.interpretation_message,
                answer.dispatch_message,
            ]
            if message is not None
        ]
        history.extend(ai_messages)

        if is_task_done:
            break

        agent = (
            agent_registry[dispatched_agent]
            if dispatched_agent is not None
            else start_from_agent
        )

        if epoch == max_loops:
            break

        agent.check_if_task_is_done(history)  # pyright: ignore
        epoch += 1
