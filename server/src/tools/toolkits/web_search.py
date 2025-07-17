import json

from src.services import web_search
from src.tools.tools import define_toolkit, description

tool, resource, register_toolkit = define_toolkit()


@resource.create(
    description=description(
        """
        Searches for information on the web
        IMPORTANT: Make sure you always cite your sources!
        """,
        args=[
            ("query", "The search query."),
            (
                "max_results",
                "The maximum number of results to retrieve from the search engine. Defaults to 5.",
            ),
        ],
        returns=[
            (
                "list[dict[str, str]] | None",
                'The search results in format [{ "url": <URL>, "title": <TITLE>, "content": <CONTENT> }]. Returns None if an error occured or no search results were found.',
            ),
        ],
    )
)
def search(query: str, max_results: int = 5) -> list[dict[str, str]] | None:
    query = query.strip().lower()
    return web_search.search(query, max_results)
