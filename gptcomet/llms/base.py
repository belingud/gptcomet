from abc import ABC, abstractmethod
from contextlib import contextmanager
from typing import Any, Optional, ParamSpec, TypeVar, Union

import glom
import httpx
import orjson as json
from rich.text import Text

from gptcomet.exceptions import ConfigError, ConfigErrorEnum, RequestError
from gptcomet.log import logger
from gptcomet.styles import Colors
from gptcomet.utils import console

P = ParamSpec("P")
T = TypeVar("T", bound=Any)


class BaseLLM(ABC):
    """Base class for all LLM providers."""

    def __init__(self, config: dict[str, Any]):
        """Initialize the LLM provider with configuration."""
        self.config = config
        self.api_key = str(config.get("api_key", ""))
        if not self.api_key:
            raise ConfigError(ConfigErrorEnum.API_KEY_MISSING)
        self.api_base = str(config.get("api_base", "")).rstrip("/")
        self.model = str(config.get("model", ""))
        self.retries = int(config.get("retries", 2))
        self.timeout = int(config.get("timeout", 30))
        self.proxy = str(config.get("proxy", ""))
        self.max_tokens = int(config.get("max_tokens", 100))
        self.temperature = float(config.get("temperature", 0.7))
        self.top_p = float(config.get("top_p", 1.0))
        self.frequency_penalty = config.get("frequency_penalty")
        self.presence_penalty = config.get("presence_penalty")
        self.extra_headers = str(config.get("extra_headers", "{}"))
        self.completion_path = config.get("completion_path")  # /chat/completions
        if self.completion_path:
            self.completion_path = self.completion_path.lstrip("/")
        self.answer_path = config.get("answer_path")  # choices.0.message.content
        self._client = None

    @classmethod
    def get_required_config(cls) -> dict[str, tuple[str, str]]:
        """Get provider-specific configuration requirements.

        Returns a dictionary where:
        - key: configuration field name
        - value: tuple of (default_value, prompt_message)
        """
        return {
            "api_base": ("", "Enter API Base URL"),
            "model": ("", "Enter model name"),
            "api_key": ("", "Enter API key"),
            "max_tokens": ("1024", "Enter max tokens"),
        }

    @property
    def client(self) -> httpx.Client:
        """Lazy initialization of HTTP client."""
        if self._client is None:
            limits = httpx.Limits(max_keepalive_connections=5, max_connections=10)
            transport = httpx.HTTPTransport(retries=self.retries, limits=limits)
            client_params = {"transport": transport, "timeout": self.timeout}
            if self.proxy:
                client_params["proxy"] = self.proxy
            self._client = httpx.Client(**client_params)
        return self._client

    @contextmanager
    def managed_client(self):
        """Context manager for HTTP client."""
        try:
            yield self.client
        finally:
            if self._client is not None:
                self._client.close()
                self._client = None

    @abstractmethod
    def format_messages(
        self, message: str, history: Optional[list[dict[str, str]]] = None
    ) -> dict[str, Any]:
        """Format messages for the provider's API."""
        pass

    def build_url(self) -> str:
        """Build the API URL."""
        return f"{self.api_base}/{self.completion_path}"

    def build_headers(self) -> dict[str, str]:
        """Build request headers."""
        default_headers = {
            "Content-Type": "application/json",
            "Authorization": f"Bearer {self.api_key}",
        }
        return {**default_headers, **(json.loads(self.config.get("extra_headers", "{}")))}

    def parse_response(self, response: dict[str, Any]) -> str:
        """Parse the response from the API."""
        text = glom.glom(response, self.answer_path)
        if isinstance(text, str):
            if text.startswith("```") and text.endswith("```"):
                return text[3:-3].strip()
            return text.strip()
        return text

    def get_usage(self, data: dict[str, Any]) -> Optional[Union[str, Text]]:
        """Print usage information for the provider."""
        usage = data.get("usage")
        if not usage:
            return None
        else:
            text = Text("Token usage> prompt tokens: ")
            text.append(f"{usage.get('prompt_tokens')}", Colors.LIGHT_MAGENTA_RGB)
            text.append(", completion tokens: ")
            text.append(f"{usage.get('completion_tokens')}", Colors.LIGHT_MAGENTA_RGB)
            text.append(" total tokens: ")
            text.append(f"{usage.get('total_tokens')}", Colors.LIGHT_MAGENTA_RGB)
            return text

    def make_request(
        self, message: str, history: Optional[list[dict[str, str]]] = None, **kwargs
    ) -> str:
        """Make a request to the API."""
        url = self.build_url()
        headers = self.build_headers()
        payload = self.format_messages(message, history)

        try:
            with self.managed_client() as client:
                response = client.post(url, json=payload, headers=headers)
                logger.debug(f"Request URL: {url}")
                logger.debug(f"Response: {response.json()}")
                response.raise_for_status()
                data = response.json()
                usage = self.get_usage(data)
                if usage:
                    console.print(usage)
                return self.parse_response(data)
        except httpx.TimeoutException as e:
            msg = f"Request timed out: {e}"
            raise RequestError(msg) from e
        except httpx.HTTPStatusError as e:
            msg = f"Request failed with status code {e.response.status_code}: {e}"
            raise RequestError(msg) from e
        except Exception as e:
            msg = f"Request failed with error: {e}"
            raise RequestError(msg) from e

    def __del__(self):
        """Cleanup method to close HTTP client."""
        if hasattr(self, "_client") and self._client is not None:
            self._client.close()

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        if hasattr(self, "_client") and self._client is not None:
            self._client.close()
            self._client = None
