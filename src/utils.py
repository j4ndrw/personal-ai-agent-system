from src.client import ollama_unsafe_client


def load_model(*, model: str):
    models = ollama_unsafe_client.list().models
    for found in models:
        if found.model == model:
            return

    ollama_unsafe_client.pull(model)
