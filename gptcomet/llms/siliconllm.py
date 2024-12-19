from typing import Any

from gptcomet.llms.openai import OpenaiLLM


class SiliconLLM(OpenaiLLM):
    """Silicon LLM provider implementation."""

    def __init__(self, config: dict[str, Any]):
        super().__init__(config)

        self.api_base = config.get("api_base") or "https://api.siliconflow.cn/v1"
        self.model = config.get("model") or "silicon-1"
        self.completion_path = config.get("completion_path") or "chat/completions"
        self.answer_path = config.get("answer_path") or "choices.0.message.content"

    @classmethod
    def get_required_config(cls) -> dict[str, tuple[str, str]]:
        return {
            "api_base": ("https://api.siliconflow.cn/v1", "Enter Silicon API Base URL"),
            "model": ("Qwen/Qwen2.5-7B-Instruct", "Enter model name"),
            "api_key": ("", "Enter API key"),
            "max_tokens": ("1024", "Enter max tokens"),
        }
