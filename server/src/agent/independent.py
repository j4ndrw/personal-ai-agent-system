from ollama import Message
from src.constants import INTERPRETATION_MODEL, NON_AGENTIC_MODEL
from src.prompts import (summarizer_system_message,
                         web_summarizer_system_message)
from src.services.chat import create_chat_handler, create_chat_handler_no_stream


def summarization_independent_agent(history_chunk: list[Message]):
    user_message, assistant_messages = history_chunk[0], history_chunk[1:]
    chat = create_chat_handler()
    stream = chat(
        history=[summarizer_system_message(), user_message, *assistant_messages],
        model=NON_AGENTIC_MODEL,
        with_tools=False,
        think=False,
        mark_as_thinking=True
    )
    for token in stream:
        yield token

    summary = stream.ret
    return [user_message, summary]

def web_summarization_independent_agent(sources: list[tuple[str, str, str]]) -> str:
    chat = create_chat_handler_no_stream()
    message = chat(
        history=[
            web_summarizer_system_message(),
            *[
                Message(role="assistant", content=f"Source URL: {url}. Title: {title}. Content: {content}")
                for url, title, content in sources
            ]
        ],
        model=INTERPRETATION_MODEL,
        with_tools=False,
        think=False,
    )
    return message.content or ""
