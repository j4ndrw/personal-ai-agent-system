import json

import ollama
from fastapi import FastAPI, Response
from fastapi.middleware.cors import CORSMiddleware

from src.agent.agent import agent_registry
from src.history import history
from src.models.requests.chat import Chat
from src.prompts import system_message
from src.agent.agents import master_agent

app = FastAPI()
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

@app.post("/api/chat")
async def chat(request: Chat):
    if len(history) == 0:
        history.append(system_message)
    ephemeral_history: list[ollama.Message] = []

    print(f"USER: {request.user_prompt}")
    user_message = ollama.Message(role="user", content=request.user_prompt)
    history.append(user_message)
    ephemeral_history.append(user_message)

    agent = master_agent
    while True:
        ai_messages, dispatched_agent = agent(history)
        history.extend(ai_messages)
        ephemeral_history.extend(ai_messages)

        if dispatched_agent is None:
            break

        print(f"`{agent.name}` agent delegated action to `{dispatched_agent}` agent...") # pyright: ignore
        agent = agent_registry[dispatched_agent]

    return Response(
        json.dumps(
            [
                {"role": message.role, "content": message.content}
                for message in ephemeral_history
            ]
        ),
        media_type="text/plain",
    )
