from src.agent.agent import register_agent

help_agent = register_agent(
    name="help",
    when_to_dispatch="user asks the AI what its capabilities are",
    toolkits=["help"],
)
chat_agent = register_agent(
    name="chat",
    when_to_dispatch="user chats regularly, without asking for information or for tasks to be performed",
)
web_search_agent = register_agent(
    name="web_search",
    when_to_dispatch="user asks for information on something",
    toolkits=["web_search"],
)
router_agent = register_agent(name="router", toolkits=["router"])
