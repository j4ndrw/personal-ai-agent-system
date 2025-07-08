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
    ephemeral_history: list[ollama.Message] = []

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
                answer.tool_result_message,
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
        json.dumps(
            [
                {
                    "agentic_message": (
                        answer.agentic_message.content
                        if answer.agentic_message is not None
                        else None
                    ),
                    "non_agentic_message": (
                        answer.non_agentic_message.content
                        if answer.non_agentic_message is not None
                        else None
                    ),
                    "tool_result_message": (
                        answer.tool_result_message.content
                        if answer.tool_result_message is not None
                        else None
                    ),
                    "interpretation_message": (
                        answer.interpretation_message.content
                        if answer.interpretation_message is not None
                        else None
                    ),
                    "dispatch_message": (
                        answer.dispatch_message.content
                        if answer.dispatch_message is not None
                        else None
                    ),
                }
                for answer in answers
            ]
        ),
        media_type="text/plain",
    )
