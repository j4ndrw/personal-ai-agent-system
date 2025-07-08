from src.tools.tools import define_toolkit, description

_, resource, register_toolkit = define_toolkit()


@resource.create(
    description=description(
        """
        Provides the user information regarding what tools are available
        """,
        returns=[
            (
                "list[str]",
                "The list of things that the AI can do.",
            )
        ],
    )
)
def help() -> list[str]:
    return [
        "@web <QUERY> - Searches for something on the web.",
        "@utility #delete-conversation - Deletes the current conversation with the AI"
    ]
