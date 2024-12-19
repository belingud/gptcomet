from typing import Any

from .openai import OpenaiLLM


class DeepseekLLM(OpenaiLLM):
    def __init__(self, config: dict[str, Any]):
        super().__init__(config)
        self.api_base = config.get("api_base") or "https://api.deepseek.com/beta"
        self.model = config.get("model") or "deepseek-chat"
        self.completion_path = config.get("completion_path") or "completions"
        self.answer_path = config.get("answer_path") or "choices.0.message.content"

    @classmethod
    def get_required_config(cls) -> dict[str, tuple[str, str]]:
        return {
            "api_base": ("https://api.deepseek.com/beta", "Enter DeepSeek API Base URL"),
            "model": ("deepseek-chat", "Enter model name"),
            "api_key": ("", "Enter API key"),
            "max_tokens": ("1024", "Enter max tokens"),
        }
