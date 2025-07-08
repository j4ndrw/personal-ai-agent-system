import json
import ollama
from fastapi import FastAPI, Response
from fastapi.middleware.cors import CORSMiddleware

from src.agent.agent import register_agent
from src.constants import (CHAT_MODEL, HELP_MODEL, MASTER_MODEL,
                           WEB_SEARCH_MODEL)
from src.history import history
from src.models.requests.chat import Chat

app = FastAPI()
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

help_agent = register_agent(
    name="help",
    when_to_dispatch="user asks the AI what its capabilities are",
    model=HELP_MODEL,
    toolkits=["help"]
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


@app.post("/api/chat")
async def chat(request: Chat):
    ephemeral_history: list[ollama.Message] = []

    if request.system_prompt:
        print(f"SYSTEM: {request.system_prompt}")
        system_message = ollama.Message(role="system", content=request.system_prompt)
        history.append(system_message)

    print(f"USER: {request.user_prompt}")
    user_message = ollama.Message(role="user", content=request.user_prompt)
    history.append(user_message)
    ephemeral_history.append(user_message)

    ai_messages = master_agent(history)
    history.extend(ai_messages)
    ephemeral_history.extend(ai_messages)

    return Response(
        json.dumps(
            [
                {"role": message.role, "content": message.content}
                for message in ephemeral_history
            ]
        ),
        media_type="text/plain",
    )
