from typing import Any, Optional

from .base import BaseLLM


class VertexLLM(BaseLLM):
    """Google Vertex AI LLM provider."""

    def __init__(self, config: dict[str, Any]):
        super().__init__(config)

        self.api_base = config.get("api_base") or "https://us-central1-aiplatform.googleapis.com/v1"
        self.model = config.get("model") or "gemini-pro"
        self.project_id = config.get("project_id")
        self.location = config.get("location") or "us-central1"

        if not self.project_id:
            msg = "project_id is required for Vertex AI"
            raise ValueError(msg)

        # Construct the model path
        self.completion_path = config.get("completion_path") or (
            f"projects/{self.project_id}/locations/{self.location}/publishers/google/models/{self.model}:generateContent"
        )
        self.answer_path = config.get("answer_path") or "candidates.0.content.parts.0.text"

    def build_headers(self) -> dict[str, str]:
        """Build request headers."""
        return {
            "Content-Type": "application/json",
            "Authorization": f"Bearer {self.api_key}",
        }

    def format_messages(self, message: str, history: Optional[list[dict[str, str]]] = None) -> dict:
        """Build request payload."""
        messages = []

        if history:
            for msg in history:
                messages.append(
                    {
                        "author": msg["role"],
                        "content": msg["content"],
                    }
                )

        messages.append(
            {
                "author": "user",
                "content": message,
            }
        )

        payload = {
            "contents": messages,
            "generation_config": {
                "max_output_tokens": self.max_tokens,
                "temperature": self.temperature,
                "top_p": self.top_p,
            },
        }

        if self.frequency_penalty is not None:
            payload["generation_config"]["frequency_penalty"] = float(self.frequency_penalty)

        if self.presence_penalty is not None:
            payload["generation_config"]["presence_penalty"] = float(self.presence_penalty)

        return payload

    @classmethod
    def get_required_config(cls) -> dict[str, tuple[str, str]]:
        """Get Vertex AI-specific configuration requirements."""
        return {
            "api_base": (
                "https://us-central1-aiplatform.googleapis.com/v1",
                "Enter Vertex AI API Base URL",
            ),
            "model": ("gemini-pro", "Enter model name"),
            "api_key": ("", "Enter API key"),
            "project_id": ("", "Enter Google Cloud project ID"),
            "location": ("us-central1", "Enter location (e.g., us-central1)"),
            "max_tokens": ("1024", "Enter max tokens"),
        }
