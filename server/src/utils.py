from typing import Any, Generator, Generic, TypeVar

import readability
from bs4 import BeautifulSoup
from src.client import ollama_unsafe_client
from src.models.agent.answer import Answer


def load_model(*, model: str):
    models = ollama_unsafe_client.list().models
    for found in models:
        if found.model == model:
            return

    ollama_unsafe_client.pull(model)


def prettify_xml(xml: str):
    bs = BeautifulSoup(xml, "xml")
    return str(bs.decode(indent_level=4))


def combined_response(answer: Answer) -> str | None:
    non_agentic_message = (
        answer.non_agentic_message.content
        if answer.non_agentic_message is not None
        else None
    )
    tools = answer.tool_result_message.items()
    tool_result_message = "\n---\n".join(
        [
            f"""
`{tool}` tool:
```json
{tool_result.content}
```
"""
            for tool, tool_result in tools
            if tool_result.content is not None
        ]
    )
    interpretation_message = (
        answer.interpretation_message.content
        if answer.interpretation_message is not None
        else None
    )

    if not interpretation_message and not non_agentic_message and len(tools) == 0:
        return None

    if not interpretation_message or len(answer.tool_result_message.items()) == 0:
        return non_agentic_message

    return f"""
Tools used:
{tool_result_message}
---
{interpretation_message}
"""


def html_to_text(html: str) -> str:
    doc = readability.Document(html)
    parsed_html = doc.summary()

    soup = BeautifulSoup(parsed_html, "html.parser")
    text = soup.get_text(separator=" ", strip=True)

    return text


TYield = TypeVar("TYield")
TReturn = TypeVar("TReturn")


class StatefulGenerator(Generic[TYield, TReturn]):
    def __init__(self, g: Generator[TYield, Any, TReturn]):
        self.g = g
        self.ret: TReturn = None  # pyright: ignore

    def __iter__(self):
        self.ret = yield from self.g

def truncate(s: str, max_chars = 100):
    if len(s) > max_chars:
        return f"{s[:max_chars]}[...]"
    return s
