from ddgs import DDGS

ddgs = DDGS()


def search_on_duckduckgo(query: str, max_results: int) -> list[dict[str, str]]:
    try:
        results = ddgs.text(query, max_results=max_results)
        filtered_results = [result for result in results if "body" in result]
        return filtered_results
    except Exception as e:
        print(f"Error searching internet: {e}")
        return []
