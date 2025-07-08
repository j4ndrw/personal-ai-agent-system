import json

from src.services import web_search
from src.tools.tools import define_toolkit, description

tool, resource, register_toolkit = define_toolkit()


@resource.create(
    description=description(
        """
        Searches for information using the DuckDuckGo search engine.
        IMPORTANT: Only use this tool if the user prepends "@web" in the prompt.
        IMPORTANT: Make sure you always cite your sources!
        """,
        args=[
            ("query", "The search query."),
            (
                "max_results",
                "The maximum number of results to retrieve from the search engine. Defaults to 10.",
            ),
        ],
        returns=[
            (
                "str | None",
                "The search results. Returns None if an error occured or no search results were found.",
            ),
            (
                "str | None",
                'The sources to be cited, in the following JSON format: `[{ "title": "Lorem ipsum", "url": "lipsum.com", "snippet": "Lorem ipsum dolor sit amet" }]`. Returns None if an error occured or no search results were found.',
            ),
        ],
    )
)
def search_on_duckduckgo(
    query: str, max_results: int = 10
) -> tuple[str | None, str | None]:
    query = query.strip().lower()
    search_results = web_search.search_on_duckduckgo(query, max_results)
    search_content = (
        "\n".join([result["body"] for result in search_results])
        if len(search_results) > 0
        else None
    )
    sources = (
        json.dumps(
            [
                {
                    "title": result["title"],
                    "url": result["href"],
                    "snippet": result["body"],
                }
                for result in search_results
            ]
        )
        if len(search_results) > 0
        else None
    )
    return search_content, sources
