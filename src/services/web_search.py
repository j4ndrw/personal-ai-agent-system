import requests
from ddgs import DDGS
from readability.readability import urllib

from src.constants import SEARXNG_HOST
from src.utils import html_to_text

ddgs = DDGS()


def search_on_duckduckgo(query: str, max_results: int) -> list[dict[str, str]]:
    try:
        results = ddgs.text(query, max_results=max_results)
        filtered_results = [result for result in results if "body" in result]
        return filtered_results
    except Exception as e:
        print(f"Error searching internet: {e}")
        return []


def search(query: str, max_results: int) -> list[dict[str, str]] | None:
    query = urllib.parse.quote(query)

    spoofed_user_agent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
    response = requests.get(
        f"{SEARXNG_HOST}/search?q={query}&format=json",
        headers={"User-Agent": spoofed_user_agent},
    )
    if response.status_code != 200:
        print("Huh?", response)
        return None

    results = response.json().get("results", [])

    contexts: list[dict[str, str]] = []
    visited_index = 0
    while (
        len(contexts) < min(max_results, len(results))
        and visited_index < len(results)
    ):
        if len(results) == 0:
            break

        result = results[visited_index]
        visited_index += 1

        try:
            search_response = requests.get(result["url"])
            if search_response.status_code != 200:
                continue

            html = search_response.text
            contexts.append({"url": result["url"], "content": html_to_text(html)})
        except Exception:
            continue

    return contexts
