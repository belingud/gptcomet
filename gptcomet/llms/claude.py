from typing import Any, Optional

from .base import BaseLLM


class ClaudeLLM(BaseLLM):
    def __init__(self, config: dict[str, Any]):
        super().__init__(config)

        self.api_base = config.get("api_base") or "https://api.anthropic.com/v1"
        self.model = config.get("model") or "claude-3-5-sonnet"
        self.completion_path = config.get("completion_path") or "messages"
        self.answer_path = config.get("answer_path") or "content.0.text"
        self.anthropic_version = config.get("anthropic-version", "2023-06-01")
        self.top_k = config.get("top_k")

    def build_headers(self):
        headers = {
            "x-api-key": self.api_key,
            "Content-Type": "application/json",
            "anthropic-version": self.anthropic_version,
        }
        return headers

    def format_messages(
        self, message: str, history: Optional[list[dict[str, str]]] = None
    ) -> dict[str, Any]:
        payload = {
            "model": self.model,
            "prompt": message,
            "max_tokens": self.max_tokens,
        }
        if self.top_k:
            payload["top_k"] = self.top_k

        if self.temperature is not None:
            payload["temperature"] = float(self.temperature)

        if self.top_p is not None:
            payload["top_p"] = float(self.top_p)

        return payload

    def get_usage(self, data: dict[str, Any]) -> Optional[str]:
        usage = data.get("usage")
        if not usage:
            return None
        return (
            f"Token usage: input tokens: {usage.get('input_tokens')}, "
            f"output tokens: {usage.get('output_tokens')}, "
        )

    @classmethod
    def get_required_config(cls) -> dict[str, tuple[str, str]]:
        return {
            "api_base": ("https://api.anthropic.com/v1", "Enter Anthropic API Base URL"),
            "model": ("claude-3-5-sonnet", "Enter model name"),
            "api_key": ("", "Enter API key"),
            "max_tokens": ("1024", "Enter max tokens"),
            "anthropic-version": ("2023-06-01", "Enter Anthropic API version"),
        }
