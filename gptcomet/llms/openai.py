from typing import Any, Optional

from gptcomet.llms.base import BaseLLM


class OpenaiLLM(BaseLLM):
    """OpenAI LLM provider implementation."""

    def __init__(self, config: dict[str, Any]):
        super().__init__(config)

        self.api_base = config.get("api_base") or "https://api.openai.com/v1"
        self.model = config.get("model") or "gpt-4o"
        self.completion_path = self.completion_path or "chat/completions"
        self.answer_path = self.answer_path or "choices.0.message.content"

    def build_headers(self) -> dict[str, str]:
        """Build request headers."""
        return {
            "Content-Type": "application/json",
            "Authorization": f"Bearer {self.api_key}",
        }

    def format_messages(
        self, message: str, history: Optional[list[dict[str, str]]] = None
    ) -> dict[str, Any]:
        """Format messages for OpenAI API."""
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

        if self.top_p is not None:
            payload["top_p"] = float(self.top_p)

        if self.frequency_penalty is not None:
            payload["frequency_penalty"] = float(self.frequency_penalty)

        if self.presence_penalty is not None:
            payload["presence_penalty"] = float(self.presence_penalty)

        return payload

    @classmethod
    def get_required_config(cls) -> dict[str, tuple[str, str]]:
        """Get OpenAI-specific configuration requirements."""
        return {
            "api_base": ("https://api.openai.com/v1", "Enter OpenAI API Base URL"),
            "model": ("gpt-4o", "Enter model name"),
            "api_key": ("", "Enter API key"),
            "max_tokens": ("1024", "Enter max tokens"),
        }
