from typing import Any

from .openai import OpenaiLLM


class ChatGLMLLM(OpenaiLLM):
    """ChatGLM LLM provider implementation."""

    def __init__(self, config: dict[str, Any]):
        super().__init__(config)
        self.api_base = config.get("api_base") or "https://open.bigmodel.cn/api/paas/v4"
        self.model = config.get("model") or "glm-4-flash"

    @classmethod
    def get_required_config(cls):
        return {
            "api_base": ("https://open.bigmodel.cn/api/paas/v4", "Enter ChatGLM API Base URL"),
            "model": ("glm-4-flash", "Enter model name"),
            "api_key": ("", "Enter API key"),
            "max_tokens": ("1024", "Enter max tokens"),
        }
