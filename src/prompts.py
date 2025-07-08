import ollama

from src.agent.agent import agent_registry

available_agents = "\n".join(
    [f"- {agent}" for agent in agent_registry.keys() if agent != "master"]
)

agent_instructions = "\n".join(
    [
        f"<{agent}-agent>The master agent should dispatch the {agent} agent when {agent_registry[agent].when_to_dispatch}</{agent}-agent>"  # pyright: ignore
        for agent in agent_registry.keys()
        if agent != "master"
    ]
)
master_agent_instructions = f"<master-agent>Use this agent to take the user's prompt and delegate a different agent to fulfill the task or respond to the query.</master-agent>"


SYSTEM_PROMPT = (
    lambda: f"""
You are an agentic AI application that can chat with the user,
as well as perform tasks and provide information from reliable sources.

You have the following agents available:

- master
{available_agents}

{master_agent_instructions}
{agent_instructions}
"""
)

system_message = lambda: ollama.Message(role="system", content=SYSTEM_PROMPT())
