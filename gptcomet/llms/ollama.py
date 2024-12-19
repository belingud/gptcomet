from typing import Any, Optional

from .base import BaseLLM


class OllamaLLM(BaseLLM):
    """Ollama LLM provider implementation."""

    def __init__(self, config: dict[str, Any]):
        super().__init__(config)

        self.api_base = config.get("api_base") or "http://localhost:11434/api"
        self.model = config.get("model") or "llama2"
        self.completion_path = config.get("completion_path") or "generate"
        self.answer_path = config.get("answer_path") or "response"
        self.seed = self.config.get("seed")
        self.top_k = self.config.get("top_k")
        self.num_gpu = self.config.get("num_gpu")
        self.main_gpu = self.config.get("main_gpu")
        self.repetition_penalty = self.config.get("repetition_penalty")

    def format_messages(
        self, message: str, history: Optional[list[dict[str, str]]] = None
    ) -> dict[str, Any]:
        """Format messages for Ollama API."""

        payload = {
            "model": self.model,
            "prompt": message,
            "options": {
                "num_predict": self.max_tokens,
            },
        }

        if self.temperature is not None:
            payload["options"]["temperature"] = float(self.temperature)
        if self.repetition_penalty is not None:
            payload["options"]["repetition_penalty"] = float(self.repetition_penalty)
        if self.presence_penalty is not None:
            payload["options"]["presence_penalty"] = float(self.presence_penalty)
        if self.frequency_penalty is not None:
            payload["options"]["frequency_penalty"] = float(self.frequency_penalty)
        if self.top_p is not None:
            payload["options"]["top_p"] = float(self.top_p)
        if self.top_k is not None:
            payload["options"]["top_k"] = int(self.top_k)
        if self.seed is not None:
            payload["options"]["seed"] = int(self.seed)
        if self.num_gpu is not None:
            payload["options"]["num_gpu"] = int(self.num_gpu)
        if self.main_gpu is not None:
            payload["options"]["main_gpu"] = int(self.main_gpu)

        return payload

    @classmethod
    def get_required_config(cls) -> dict[str, tuple[str, str]]:
        return {
            "api_base": ("http://localhost:11434/api", "Enter Ollama API Base URL"),
            "model": ("llama2", "Enter model name"),
            "max_tokens": ("1024", "Enter max tokens"),
        }
