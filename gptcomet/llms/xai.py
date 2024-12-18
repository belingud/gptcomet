from typing import Any

from .openai import OpenaiLLM


class XaiLLM(OpenaiLLM):
    """XAI (讯飞星火) LLM provider implementation."""

    def __init__(self, config: dict[str, Any]):
        super().__init__(config)

        self.api_base = self.api_base or "https://api.x.ai/v1/"
        self.model = self.model or "grok-beta"
