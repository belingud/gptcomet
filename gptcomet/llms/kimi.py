from typing import Any

from .openai import OpenaiLLM


class KimiLLM(OpenaiLLM):
    def __init__(self, config: dict[str, Any]):
        super().__init__(config)

        self.api_base = config.get("api_base") or "https://api.kimi.ai/v1"
        self.model = config.get("model") or "moonshot-v1-8k"
        self.completion_path = self.completion_path or "chat/completions"
        self.answer_path = self.answer_path or "choices.0.message.content"

    @classmethod
    def get_required_config(cls) -> dict[str, tuple[str, str]]:
        return {
            "api_base": ("https://api.kimi.ai/v1", "Enter Kimi API Base URL"),
            "model": ("moonshot-v1-8k", "Enter model name"),
            "api_key": ("", "Enter API key"),
            "max_tokens": ("1024", "Enter max tokens"),
        }
