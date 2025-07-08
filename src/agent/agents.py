from typing import Callable

from ollama import Message
from src.agent.agent import register_agent
from src.constants import (CHAT_MODEL, HELP_MODEL, MASTER_MODEL,
                           WEB_SEARCH_MODEL)


help_agent = register_agent(
    name="help",
    when_to_dispatch="user asks the AI what its capabilities are",
    model=HELP_MODEL,
    toolkits=["help"],
)
chat_agent = register_agent(
    name="chat",
    when_to_dispatch="user chats regularly, without asking for information or for tasks to be performed",
    model=CHAT_MODEL,
)
web_search_agent = register_agent(
    name="web_search",
    when_to_dispatch='user asks for information on something - requires "@web" prefix',
    model=WEB_SEARCH_MODEL,
    toolkits=["web_search"],
)
utility_agent = register_agent(
    name="utility",
    when_to_dispatch='user wants to perform a utility action - requires "@utility" prefix',
    model=WEB_SEARCH_MODEL,
    toolkits=["utility"],
)
master_agent = register_agent(
    name="master", model=MASTER_MODEL, toolkits=["dispatch_agent"]
)
