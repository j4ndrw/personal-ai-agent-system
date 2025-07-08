from pydantic import BaseModel


class Chat(BaseModel):
    user_prompt: str
