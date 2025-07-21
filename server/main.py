import json
import ollama
from fastapi import FastAPI, Response
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import StreamingResponse
from src.agent.independent import simple_independent_agent
from src.agent.agent import agentic_loop
from src.agent.agents import router_agent
from src.agent.registry import agent_registry
from src.history import history
from src.models.requests.chat import Chat, ChatWithAgent

app = FastAPI()
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


@app.post("/api/agentic/auto")
async def agentic_auto(request: Chat):
    user_message = ollama.Message(role="user", content=request.prompt)
    history.append(user_message)

    return StreamingResponse(
        agentic_loop(
            history,
            start_from_agent=router_agent,
        ),
        media_type="text/event-stream",
    )

@app.post("/api/agentic/manual")
async def agentic_manual(request: ChatWithAgent):
    user_message = ollama.Message(role="user", content=request.prompt)
    history.append(user_message)

    return StreamingResponse(
        agentic_loop(
            history,
            start_from_agent=agent_registry[request.agent],
        ),
        media_type="text/event-stream",
    )

@app.post("/api/simple")
async def simple(request: Chat):
    if len(history) > 0 and history[0].role == "system":
        history.pop(0)

    user_message = ollama.Message(role="user", content=request.prompt)
    history.append(user_message)

    return StreamingResponse(
        simple_independent_agent(history),
        media_type="text/event-stream",
    )

@app.get("/api/agents")
async def agents():
    return Response(json.dumps(list(agent_registry.keys())), media_type="application/json")
