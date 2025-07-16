from pydantic import BaseModel


class Chat(BaseModel):
    prompt: str
