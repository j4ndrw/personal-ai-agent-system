import ollama
from src.agent.agent import agent_registry

available_agents = "\n".join(
    [f"- {agent}" for agent in agent_registry.keys() if agent != "router"]
)

agent_instructions = "\n".join(
    [
        f"<{agent}-agent>The router agent should dispatch the {agent} agent when {agent_registry[agent].when_to_dispatch}</{agent}-agent>"  # pyright: ignore
        for agent in agent_registry.keys()
        if agent != "router"
    ]
)
router_agent_instructions = f"<router-agent>Use this agent to take the user's prompt and delegate a different agent to fulfill the task or respond to the query.</router-agent>"


SYSTEM_PROMPT = (
    lambda: f"""
You are an agentic AI application that can chat with the user,
as well as perform tasks and provide information from reliable sources.

You have the following agents available:

- router
{available_agents}

{router_agent_instructions}
{agent_instructions}
"""
)

system_message = lambda: ollama.Message(role="system", content=SYSTEM_PROMPT())
