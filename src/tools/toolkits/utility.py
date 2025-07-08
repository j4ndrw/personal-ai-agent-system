from src.agent.agent import agents
from src.tools.tools import define_toolkit, description

tool, _, register_toolkit = define_toolkit()


@tool.create(
    description=description(
        """
        Clears the conversation history and starts a fresh one.
        IMPORTANT: Only use this tool if the user prepends "@utility #delete-conversation" in their prompt.
        """,
        returns=[
            (
                "str | None",
                "The agent to dispatch. If the agent is not found or invalid, returns None.",
            )
        ],
    )
)
def dispatch_agent(agent: str) -> str | None:
    if agent not in agents.keys() and agent != "master":
        return None

    return agent
