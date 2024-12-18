from typing import Any

from .openai import OpenaiLLM


class SambanovaLLM(OpenaiLLM):
    """Sambanova LLM provider implementation."""

    def __init__(self, config: dict[str, Any]):
        super().__init__(config)

        self.api_base = self.api_base or "https://api.sambanova.ai/v1"
        self.model = self.model or "Meta-Llama-3.3-70B-Instruct"
