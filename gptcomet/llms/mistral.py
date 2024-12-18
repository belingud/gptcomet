from typing import Any

from .openai import OpenaiLLM


class MistralLLM(OpenaiLLM):
    """Mistral AI LLM provider implementation."""

    def __init__(self, config: dict[str, Any]):
        super().__init__(config)

        self.api_base = self.api_base or "https://api.mistral.ai/v1"
        self.model = self.model or "mistral-large-latest"
