import ollama
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import StreamingResponse
from src.agent.agent import agentic_loop
from src.agent.agents import router_agent
from src.history import history
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

    user_message = ollama.Message(role="user", content=request.prompt)
    history.append(user_message)

    return StreamingResponse(
        agentic_loop(
            history,
            start_from_agent=router_agent,
        ),
        media_type="text/event-stream",
    )
