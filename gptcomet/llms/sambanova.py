from typing import Any

from .openai import OpenaiLLM


class SambanovaLLM(OpenaiLLM):
    """Sambanova LLM provider implementation."""

    def __init__(self, config: dict[str, Any]):
        super().__init__(config)

        self.api_base = config.get("api_base") or "https://api.sambanova.ai/v1"
        self.model = config.get("model") or "Meta-Llama-3.3-70B-Instruct"

    @classmethod
    def get_required_config(cls) -> dict[str, tuple[str, str]]:
        return {
            "api_base": ("https://api.sambanova.ai/v1", "Enter Sambanova API Base URL"),
            "model": ("Meta-Llama-3.3-70B-Instruct", "Enter model name"),
            "api_key": ("", "Enter API key"),
            "max_tokens": ("1024", "Enter max tokens"),
        }
