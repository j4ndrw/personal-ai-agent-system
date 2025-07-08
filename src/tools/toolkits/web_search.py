from src.tools.tools import define_toolkit, description

tool, resource, register_toolkit = define_toolkit()


@resource.create(
    description=description(
        """
        Searches for information using the DuckDuckGo search engine.
        IMPORTANT: Only use this tool if the user prepends "@web" in the prompt.
        """,
        args=[
            ("query", "The search query."),
            (
                "max_results",
                "The maximum number of results to retrieve from the search engine. Defaults to 10.",
            ),
        ],
    )
)
def search_on_duckduckgo(query: str, max_results: int = 10):
    pass
