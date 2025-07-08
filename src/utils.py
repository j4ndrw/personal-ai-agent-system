from bs4 import BeautifulSoup

from src.client import ollama_unsafe_client


def load_model(*, model: str):
    models = ollama_unsafe_client.list().models
    for found in models:
        if found.model == model:
            return

    ollama_unsafe_client.pull(model)


def prettify_xml(xml: str):
    bs = BeautifulSoup(xml, "xml")
    return str(bs.decode(indent_level=4))
