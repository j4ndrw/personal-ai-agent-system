from typing import Any

from src.services import web_search
from src.tools.tools import define_toolkit, description

tool, resource, register_toolkit = define_toolkit()


@resource.create(
    description=description(
        """
        Searches for information on the web
        IMPORTANT: Make sure you always cite your sources!
        IMPORTANT: Before calling the tool, make sure to rephrase the query so
        that you get better search results. Also, adjust the `max_results` parameter
        as needed to get the task done - this is a parameter you can control and doesn't have
        to be left as the default.
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
                "dict[str, Any] | None",
                'The search results in format [{ "sources": [{ "url": <URL>, "title": <TITLE> }], content: <SUMMARIZED_CONTENT> }]. Returns None if an error occured or no search results were found.',
            ),
        ],
    )
)
def search(query: str, max_results: int = 5) -> dict[str, Any] | None:
    query = query.strip().lower()
    return web_search.search(query, max_results)
