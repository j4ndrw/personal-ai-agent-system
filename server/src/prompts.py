import ollama
from src.agent.registry import agent_registry

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

MULTI_AGENT_SYSTEM_PROMPT = (
    lambda name: f"""
You are an agentic AI application that can chat with the user,
as well as perform tasks and provide information from reliable sources.

You are currently using the {name} agent.

You have the following agents available:

- router
{available_agents}

{router_agent_instructions}
{agent_instructions}
"""
)
multi_agent_system_message = lambda name: ollama.Message(role="system", content=MULTI_AGENT_SYSTEM_PROMPT(name))

SUMMARIZER_SYSTEM_PROMPT = (
    lambda: f"""
You are responsible with summarizing the conversation provided in the prompt.
Include details like:
- What the thought process of the assistant was
- What tool calls were used
- What those tool calls returned
- What the response was to the user's question
"""
)
summarizer_system_message = lambda: ollama.Message(role="system", content=SUMMARIZER_SYSTEM_PROMPT())

WEB_SUMMARIZER_SYSTEM_PROMPT = (
    lambda: f"""
<instructions>
    You are responsible with summarizing the content found using the web search tool.
    - Retain details concise and relevant.
    - Point out any interesting details (e.g. code snippets, historical events, places of interest, etc...)
    - Cite the source

    <important>
        - DO NOT summarize code snippets
        - DO NOT summarize step-by-step guides
        - DO NOT summarize historical events
    </important>
</instructions
"""
)
web_summarizer_system_message = lambda: ollama.Message(role="system", content=WEB_SUMMARIZER_SYSTEM_PROMPT())
