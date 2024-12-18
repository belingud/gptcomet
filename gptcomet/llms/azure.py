from typing import Any, Optional

from .openai import OpenaiLLM


class AzureLLM(OpenaiLLM):
    """Azure OpenAI LLM provider implementation."""

    def __init__(self, config: dict[str, Any]):
        super().__init__(config)

        self.api_version = config.get("api_version", "2024-02-15-preview")
        self.deployment_name = config.get("deployment_name")
        if not self.deployment_name:
            msg = "deployment_name is required for Azure OpenAI"
            raise ValueError(msg)

        self.completion_path = f"openai/deployments/{self.deployment_name}/chat/completions"

    @classmethod
    def get_required_config(cls) -> dict[str, tuple[str, str]]:
        """Get Azure-specific configuration requirements."""
        base_config = super().get_required_config()
        base_config.update({
            "deployment_name": ("", "Enter Azure OpenAI deployment name (required)"),
            "api_version": ("2024-02-15-preview", "Enter API version"),
        })
        base_config["api_base"] = ("", "Enter Azure OpenAI endpoint URL (e.g., https://YOUR_RESOURCE_NAME.openai.azure.com)")
        base_config["model"] = ("", "Enter model name (e.g., gpt-4, gpt-35-turbo)")
        return base_config

    def build_headers(self) -> dict[str, str]:
        return {
            "Content-Type": "application/json",
            "api-key": self.api_key,
        }

    def build_url(self) -> str:
        """Build the API URL."""
        return f"{self.api_base}/{self.completion_path}?api-version={self.api_version}"
