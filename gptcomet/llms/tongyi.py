from typing import Any, Optional

from .openai import OpenaiLLM


class TongyiLLM(OpenaiLLM):
    """Tongyi (通义) LLM provider implementation."""

    def __init__(self, config: dict[str, Any]):
        super().__init__(config)

        self.api_base = (
            config.get("api_base") or "https://dashscope.aliyuncs.com/compatible-mode/v1"
        )
        self.model = config.get("model") or "qwen-turbo"

    def build_headers(self):
        """Build request headers."""
        return {
            "Authorization": f"Bearer {self.api_key}",
            "Content-Type": "application/json",
        }

    def format_messages(
        self, message: str, history: Optional[list[dict[str, str]]] = None
    ) -> dict[str, Any]:
        """Format messages for Tongyi API."""
        messages = []

        if history:
            for msg in history:
                messages.append({"role": msg["role"], "content": msg["content"]})

        messages.append({"role": "user", "content": message})

        payload = {
            "model": self.model,
            "messages": messages,
            "max_tokens": self.max_tokens,
        }

        if self.temperature is not None:
            payload["temperature"] = float(self.temperature)

        if self.frequency_penalty is not None:
            payload["frequency_penalty"] = float(self.frequency_penalty)
        if self.presence_penalty is not None:
            payload["presence_penalty"] = float(self.presence_penalty)
        if self.top_p is not None:
            payload["top_p"] = float(self.top_p)

        return payload

    @classmethod
    def get_required_config(cls) -> dict[str, tuple[str, str]]:
        return {
            "api_base": (
                "https://dashscope.aliyuncs.com/compatible-mode/v1",
                "Enter Tongyi API Base URL",
            ),
            "model": ("qwen-turbo", "Enter model name"),
            "api_key": ("", "Enter API key"),
            "max_tokens": ("1024", "Enter max tokens"),
        }
