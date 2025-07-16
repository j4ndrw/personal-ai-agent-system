from pydantic import BaseModel


class Options(BaseModel):
    temperature: int
    max_tokens: int
    stream: bool
    n: int
    presence_penalty: int
    frequency_penalty: int
    top_p: int
