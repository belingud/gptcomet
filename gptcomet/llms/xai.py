from typing import Any

from .openai import OpenaiLLM


class XaiLLM(OpenaiLLM):
    """XAI (讯飞星火) LLM provider implementation."""

    def __init__(self, config: dict[str, Any]):
        super().__init__(config)

        self.api_base = config.get("api_base") or "https://api.x.ai/v1/"
        self.model = config.get("model") or "grok-beta"

    @classmethod
    def get_required_config(cls) -> dict[str, tuple[str, str]]:
        return {
            "api_base": ("https://api.x.ai/v1/", "Enter XAI API Base URL"),
            "model": ("grok-beta", "Enter model name"),
            "api_key": ("", "Enter API key"),
            "max_tokens": ("1024", "Enter max tokens"),
        }
