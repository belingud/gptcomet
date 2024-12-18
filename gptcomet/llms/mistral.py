from typing import Any

from .openai import OpenaiLLM


class MistralLLM(OpenaiLLM):
    """Mistral AI LLM provider implementation."""

    def __init__(self, config: dict[str, Any]):
        super().__init__(config)

        self.api_base = config.get("api_base") or "https://api.mistral.ai/v1"
        self.model = config.get("model") or "mistral-large-latest"

    @classmethod
    def get_required_config(cls) -> dict[str, tuple[str, str]]:
        return {
            "api_base": ("https://api.mistral.ai/v1", "Enter Mistral API Base URL"),
            "model": ("mistral-large-latest", "Enter model name"),
            "api_key": ("", "Enter API key"),
            "max_tokens": ("1024", "Enter max tokens"),
        }
