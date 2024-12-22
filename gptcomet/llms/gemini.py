from typing import Any, Optional

from gptcomet.log import logger
from gptcomet.utils import console

from .base import BaseLLM


class GeminiLLM(BaseLLM):
    """Gemini LLM provider implementation."""

    def __init__(self, config: dict[str, Any]):
        super().__init__(config)

        self.api_base = self.api_base or "https://generativelanguage.googleapis.com/v1beta/models"
        self.model = self.model or "gemini-pro"
        self.completion_path = self.completion_path or "generateContent"
        self.answer_path = self.answer_path or "candidates.0.content.parts.0.text"

    def format_messages(
        self, message: str, history: Optional[list[dict[str, str]]] = None
    ) -> dict[str, Any]:
        """Format messages for Gemini API."""
        contents = []

        if history:
            for msg in history:
                role = "model" if msg["role"] == "assistant" else "user"
                contents.append({"role": role, "parts": [{"text": msg["content"]}]})

        contents.append({"role": "user", "parts": [{"text": message}]})

        payload = {
            "contents": contents,
            "generationConfig": {
                "maxOutputTokens": self.max_tokens,
            },
        }
        if self.temperature is not None:
            payload["generationConfig"]["temperature"] = float(self.temperature)

        if self.top_p is not None:
            payload["generationConfig"]["topP"] = float(self.top_p)

        if self.frequency_penalty is not None:
            payload["generationConfig"]["frequencyPenalty"] = float(self.frequency_penalty)

        if self.presence_penalty is not None:
            payload["generationConfig"]["presencePenalty"] = float(self.presence_penalty)

        return payload

    def build_url(self) -> str:
        """Build the API URL."""
        return f"{self.api_base}/{self.model}:generateContent?key={self.api_key}"

    def build_headers(self) -> dict[str, str]:
        """Build request headers."""
        return {
            "Content-Type": "application/json",
        }

    def make_request(
        self, message: str, history: Optional[list[dict[str, str]]] = None, **kwargs
    ) -> str:
        """Make a request to the API."""
        url = self.build_url()
        headers = self.build_headers()
        payload = self.format_messages(message, history)
        logger.debug("Sending request...")

        with self.managed_session() as session:
            response = session.post(url, json=payload, headers=headers, timeout=self.timeout)
            logger.debug(f"Request payload: {payload}")
            logger.debug(f"Response: {response.json()}")
            response.raise_for_status()
            data = response.json()
            usage = self.get_usage(data)
            if usage:
                console.print(usage)
            return self.parse_response(data)

    def get_usage(self, data: dict[str, Any]) -> Optional[str]:
        """Print usage information for the provider."""
        usage = data.get("usageMetadata")
        if not usage:
            return None
        else:
            return (
                f"Token usage: promptTokenCount: {usage.get('promptTokenCount')}, "
                f"candidatesTokenCount: {usage.get('candidatesTokenCount')}, "
                f"totalTokenCount: {usage.get('totalTokenCount')}"
            )

    @classmethod
    def get_required_config(cls) -> dict[str, tuple[str, str]]:
        return {
            "api_base": (
                "https://generativelanguage.googleapis.com/v1beta/models",
                "Enter Gemini API base",
            ),
            "model": ("gemini-pro", "Enter Gemini model"),
            "api_key": ("", "Enter Gemini API key"),
            "max_tokens": ("1024", "Enter max tokens"),
        }
