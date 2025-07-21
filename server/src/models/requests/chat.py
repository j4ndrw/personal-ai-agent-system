from pydantic import BaseModel


class Chat(BaseModel):
    prompt: str

class ChatWithAgent(BaseModel):
    prompt: str
    agent: str
