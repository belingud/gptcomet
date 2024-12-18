from typing import Any

from .openai import OpenaiLLM


class GroqLLM(OpenaiLLM):
    """Groq LLM provider implementation."""

    def __init__(self, config: dict[str, Any]):
        super().__init__(config)

        self.api_base = config.get("api_base") or "https://api.groq.com/openai/v1"
        self.model = config.get("model") or "llama3-8b-8192"

    @classmethod
    def get_required_config(cls) -> dict[str, tuple[str, str]]:
        """Get Groq-specific configuration requirements."""
        return {
            "api_base": ("https://api.groq.com/openai/v1", "Enter Groq API Base URL"),
            "model": ("llama3-8b-8192", "Enter model name"),
            "api_key": ("", "Enter API key"),
            "max_tokens": ("1024", "Enter max tokens"),
        }
