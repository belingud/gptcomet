from typing import Any, Optional

from .base import BaseLLM


class CohereLLM(BaseLLM):
    """Cohere LLM provider implementation."""

    def __init__(self, config: dict[str, Any]):
        super().__init__(config)

        self.api_base = config.get("api_base") or "https://api.cohere.ai/v1"
        self.model = config.get("model") or "command-r"
        self.answer_path = config.get("answer_path") or "text"
        self.completion_path = config.get("completion_path") or "/chat"

    def build_headers(self):
        return {
            "Authorization": f"Bearer {self.api_key}",
            "Content-Type": "application/json",
        }

    def format_messages(
        self, message: str, history: Optional[list[dict[str, str]]] = None
    ) -> dict[str, Any]:
        """Format messages for Cohere API."""
        chat_history = []

        if history:
            for msg in history:
                chat_history.append(
                    {
                        "role": "CHATBOT" if msg["role"] == "assistant" else "USER",
                        "message": msg["content"],
                    }
                )

        payload = {
            "message": message,
            "model": self.model,
            "chat_history": chat_history,
        }
        if self.max_tokens is not None:
            payload["max_tokens"] = int(self.max_tokens)
        if self.temperature is not None:
            payload["temperature"] = float(self.temperature)
        if self.top_p is not None:
            payload["top_p"] = float(self.top_p)
        if self.frequency_penalty is not None:
            payload["frequency_penalty"] = float(self.frequency_penalty)
        if self.presence_penalty is not None:
            payload["presence_penalty"] = float(self.presence_penalty)

        return payload

    def get_usage(self, data: dict[str, Any]) -> Optional[str]:
        usage = data.get("meta", {}).get("billed_units")
        if not usage:
            return None
        else:
            return f"Token usage> input tokens: {usage.get('input_tokens')}, output tokens: {usage.get('output_tokens')}"

    @classmethod
    def get_required_config(cls) -> dict[str, tuple[str, str]]:
        return {
            "api_base": ("https://api.cohere.ai/v1", "Enter Cohere API Base URL"),
            "model": ("command-r", "Enter model name"),
            "api_key": ("", "Enter API key"),
            "max_tokens": ("1024", "Enter max tokens"),
        }
