import ollama

from src.agent.agent import agentic_loop
from src.history import history
from src.prompts import system_message
from src.agent.agents import router_agent

if __name__ == "__main__":
    while True:
        prompt = input(">>> ")

        if len(history) == 0:
            history.append(system_message())

        user_message = ollama.Message(role="user", content=prompt)
        history.append(user_message)

        agentic_loop(
            history,
            start_from_agent=router_agent,
            on_token=lambda token: print(token, end="", flush=True),
        )
