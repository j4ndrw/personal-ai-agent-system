from src.agent.agent import agents
from src.tools.tools import define_toolkit, description

tool, _, register_toolkit = define_toolkit()


@tool.create(
    description=description(
        """
        Provides the user information regarding what tools are available
        IMPORTANT: Only use this tool if the user prepends "@help" in their prompt.
        """,
        returns=[
            (
                "str",
                "The list of things that the AI can do.",
            )
        ],
    )
)
def help() -> str:
    return """
        @web <QUERY> - Searches for something on the web.
        @utility #delete-conversation - Deletes the current conversation with the AI
    """
