from typing import Any, ClassVar

from .azure import AzureLLM
from .base import BaseLLM
from .chatglm import ChatGLMLLM
from .claude import ClaudeLLM
from .cohere import CohereLLM
from .deepseek import DeepseekLLM
from .gemini import GeminiLLM
from .groq import GroqLLM
from .kimi import KimiLLM
from .mistral import MistralLLM
from .ollama import OllamaLLM
from .openai import OpenaiLLM
from .sambanova import SambanovaLLM
from .siliconllm import SiliconLLM
from .tongyi import TongyiLLM
from .vertexllm import VertexLLM
from .xai import XaiLLM


class ProviderRegistry:
    """Registry for LLM providers."""

    _providers: ClassVar[dict[str, type[BaseLLM]]] = {
        "gemini": GeminiLLM,
        "cohere": CohereLLM,
        "claude": ClaudeLLM,
        "ahntropic": ClaudeLLM,
        "groq": GroqLLM,
        "mistral": MistralLLM,
        "azure": AzureLLM,
        "ollama": OllamaLLM,
        "tongyi": TongyiLLM,
        "qwen": TongyiLLM,
        "silicon": SiliconLLM,
        "chatglm": ChatGLMLLM,
        "xai": XaiLLM,
        "sambanova": SambanovaLLM,
        "openai": OpenaiLLM,
        "vertex": VertexLLM,
        "kimi": KimiLLM,
        "deepseek": DeepseekLLM,
    }

    _tongyi_default_config: ClassVar[dict[str, Any]] = {
        "api_base": "https://dashscope.aliyuncs.com/api/v1",
        "model": "qwen-max",
        "completion_path": "services/aigc/text-generation/generation",
        "answer_path": "output.text",
        "retries": 2,
        "timeout": 30,
    }
    _claude_default_config: ClassVar[dict[str, Any]] = {
        "api_base": "https://api.anthropic.com/v1",
        "model": "claude-3-5-sonnet",
        "completion_path": "messages",
        "answer_path": "content.0.text",
        "retries": 2,
        "timeout": 30,
    }

    _default_configs: ClassVar[dict[str, dict[str, Any]]] = {
        "gemini": {
            "api_base": "https://generativelanguage.googleapis.com/v1beta/models",
            "model": "gemini-pro",
            "completion_path": "generateContent",
            "answer_path": "candidates.0.content.parts.0.text",
            "retries": 2,
            "timeout": 30,
        },
        "openai": {
            "api_base": "https://api.openai.com/v1",
            "model": "gpt-3.5-turbo",
            "completion_path": "/chat/completions",
            "answer_path": "choices.0.message.content",
            "retries": 2,
            "timeout": 30,
        },
        "claude": _claude_default_config,
        "ahntropic": _claude_default_config,
        "groq": {
            "api_base": "https://api.groq.com/v1",
            "model": "mixtral-8x7b-32768",
            "completion_path": "/chat/completions",
            "answer_path": "choices.0.message.content",
            "retries": 2,
            "timeout": 30,
        },
        "mistral": {
            "api_base": "https://api.mistral.ai/v1",
            "model": "mistral-medium",
            "completion_path": "/chat/completions",
            "answer_path": "choices.0.message.content",
            "retries": 2,
            "timeout": 30,
        },
        "azure": {
            "api_base": "https://YOUR_RESOURCE_NAME.openai.azure.com",
            "deployment_name": "YOUR_DEPLOYMENT_NAME",
            "api_version": "2024-02-15-preview",
            "completion_path": "chat/completions",
            "answer_path": "choices.0.message.content",
            "retries": 2,
            "timeout": 30,
        },
        "ollama": {
            "api_base": "http://localhost:11434",
            "model": "llama2",
            "completion_path": "api/generate",
            "answer_path": "response",
            "retries": 2,
            "timeout": 30,
        },
        "tongyi": _tongyi_default_config,
        "qwen": _tongyi_default_config,
        "chatglm": {
            "api_base": "http://localhost:8000",
            "model": "glm-4-flash",
            "completion_path": "v1/chat/completions",
            "answer_path": "response",
            "retries": 2,
            "timeout": 30,
        },
        "deepseek": {
            "api_base": "https://api.deepseek.com",
            "model": "deepseek-chat",
            "completion_path": "/chat/completions",
            "answer_path": "choices.0.message.content",
            "retries": 2,
            "timeout": 30,
        },
        "openrouter": {
            "api_base": "https://api.openrouter.ai",
            "model": "meta-llama/llama-3.2-3b-instruct:free",
            "completion_path": "/chat/completions",
            "answer_path": "choices.0.message.content",
            "retries": 2,
            "timeout": 30,
        },
        "cohere": {
            "api_base": "https://api.cohere.com/v1",
            "model": "command-r-08-2024",
            "completion_path": "/chat",
            "answer_path": "text",
            "retries": 2,
            "timeout": 30,
        },
    }

    @classmethod
    def get_provider(cls, name: str) -> type[BaseLLM]:
        """Get provider class by name."""
        return cls._providers.get(name, OpenaiLLM)

    @classmethod
    def get_default_config(cls, name: str) -> dict[str, Any]:
        """Get default configuration for a provider."""
        if name not in cls._default_configs:
            return {}
        return cls._default_configs[name].copy()

    @classmethod
    def register(cls, name: str, provider_class: type[BaseLLM], default_config: dict[str, Any]):
        """Register a new provider."""
        cls._providers[name] = provider_class
        cls._default_configs[name] = default_config
