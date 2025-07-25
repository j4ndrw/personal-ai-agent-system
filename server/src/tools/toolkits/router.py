from src.agent.registry import agent_registry
from src.tools.tools import define_toolkit, description

tool, _, register_toolkit = define_toolkit()

agent_names = list(filter(lambda agent: agent != "router", agent_registry.keys()))
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
                "dict[str, str | None]",
                'The agent to dispatch. If the agent is not found or invalid, returns {"agent_to_dispatch": None}.',
            )
        ],
    )
)
def dispatch_agent(agent: str) -> dict[str, str | None]:
    if agent not in agent_registry.keys() and agent != "router":
        return {"agent_to_dispatch": None}

    return {"agent_to_dispatch": agent}


@tool.create(
    description=description(
        f"""
        Should be called if the agent thinks it finished performing the task.
        """,
        args=[],
        returns=[
            (
                "dict[str, bool]",
                'Returns {"is_task_done": True} to signify the task is completed',
            )
        ],
    )
)
def mark_task_as_done() -> dict[str, bool]:
    return {"is_task_done": True}
