from src.agent.agent import agent_registry
from src.tools.tools import define_toolkit, description

tool, _, register_toolkit = define_toolkit()

agent_names = list(filter(lambda agent: agent != "master", agent_registry.keys()))
dispatching_instructions = {
    "\n".join(
        [
            f"Dispatch {agent_name} when {agent_registry[agent_name].when_to_dispatch}"  # pyright: ignore
            for agent_name in agent_names
        ]
    )
}


@tool.create(
    description=description(
        f"""
        Dispatches an agent to take care of a task
        Can be one of the following: {agent_names}.

        <dispatching_instructions>
            {dispatching_instructions}
        </dispatching_instructions>
        """,
        args=[("agent", f"The agent to dispatch.")],
        returns=[
            (
                "str | None",
                "The agent to dispatch. If the agent is not found or invalid, returns None.",
            )
        ],
    )
)
def dispatch_agent(agent: str) -> str | None:
    print(f"Dispatching `{agent}` agent...")
    if agent not in agent_registry.keys() and agent != "master":
        return None

    return agent
