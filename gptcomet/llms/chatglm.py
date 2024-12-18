from typing import Any

from .openai import OpenaiLLM


class ChatGLMLLM(OpenaiLLM):
    """ChatGLM LLM provider implementation."""

    def __init__(self, config: dict[str, Any]):
        super().__init__(config)

        self.model = self.model or "chatglm-4-flash"
