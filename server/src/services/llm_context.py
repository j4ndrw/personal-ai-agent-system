from ollama import Message


def get_current_context_size(history: list[Message]) -> int:
    sb = ""
    for message in history:
        sb += message.content or ""
    return len(sb)
