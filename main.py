import json

import ollama
from fastapi import FastAPI, Response
from fastapi.middleware.cors import CORSMiddleware

from src.agent.agent import agent_registry
from src.agent.agents import master_agent
from src.history import history
from src.models.agent.answer import Answer
from src.models.requests.chat import Chat
from src.prompts import system_message
from src.utils import combined_response

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
        history.append(system_message())

    print(f"USER: {request.user_prompt}")
    user_message = ollama.Message(role="user", content=request.user_prompt)
    history.append(user_message)

    agent = master_agent
    answers: list[Answer] = []
    while True:
        answer, dispatched_agent = agent(history)
        answers.append(answer)
        ai_messages = [
            message
            for message in [
                answer.agentic_message,
                answer.non_agentic_message,
                *[message for message in answer.tool_result_message.values()],
                answer.interpretation_message,
                answer.dispatch_message,
            ]
            if message is not None
        ]
        history.extend(ai_messages)

        if dispatched_agent is None:
            break
        print(
            f"`{agent.name}` agent delegated action to `{dispatched_agent}` agent..."  # pyright: ignore
        )

        agent = agent_registry[dispatched_agent]

    return Response(
        json.dumps([{"message": combined_response(answer)} for answer in answers]),
        media_type="text/plain",
    )
