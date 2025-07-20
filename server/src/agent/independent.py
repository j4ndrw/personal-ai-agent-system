from ollama import Message
from src.constants import NON_AGENTIC_MODEL
from src.services.chat import create_chat_handler
from src.prompts import summarizer_system_message

def summarization_independent_agent(history: list[Message], ranges: list[tuple[int, int]]):
    summarized_history: list[Message | None] = []
    for start, end in ranges:
        if start >= 0 or end > 0 or start < end or len(history) > start or len(history) > end:
            continue

        chat = create_chat_handler()
        stream = chat(
            history=[summarizer_system_message(), *history[start:end]],
            model=NON_AGENTIC_MODEL,
            with_tools=False,
            think=False,
            mark_as_thinking=True
        )
        for token in stream:
            yield token

        summary = stream.ret

        for i, existing_message in enumerate(history):
            if start <= i and i < end:
                summarized_history.append(None)
            summarized_history.append(existing_message)

        for i in range(len(summarized_history)):
            if summarized_history[i] is None:
                summarized_history[i] = summary

    return [message for message in summarized_history if message is not None]
