import typing as t

import httpx
import orjson as json
from glom import glom
from rich.text import Text

from gptcomet.styles import Colors
from gptcomet.utils import console

try:
    import socksio  # noqa: F401
except ImportError as e:
    msg = e.msg
    if "socksio" in msg:
        msg = (
            "Using SOCKS proxy, but the 'socksio' package is not installed. "
            "Make sure to install gptcomet using `pip install gptcomet[socks]` or `pip install socksio`."
        )
    raise ImportError(msg) from None

from gptcomet._types import CompleteParams
from gptcomet.const import (
    DEFAULT_API_BASE,
    DEFAULT_MODEL,
    DEFAULT_RETRIES,
    PROVIDER_KEY,
)
from gptcomet.exceptions import ConfigError, ConfigErrorEnum
from gptcomet.log import logger

if t.TYPE_CHECKING:
    from gptcomet.config_manager import ConfigManager


class LLMClient:
    __slots__ = (
        "_http_client",
        "_request_timeout",
        "api_base",
        "api_key",
        "completion_path",
        "config_manager",
        "content_path",
        "conversation_history",
        "model",
        "provider",
        "proxy",
        "retries",
    )

    @classmethod
    def from_config_manager(cls, config_manager: "ConfigManager"):
        """
        Creates an instance of the class from a ConfigManager.

        Args:
            config_manager (ConfigManager): The ConfigManager instance to create the class instance from.

        Returns:
            An instance of the class.
        """
        return cls(config_manager)

    def __init__(self, config_manager: "ConfigManager"):
        """
        Initializes the LLMClient instance from a ConfigManager.

        Args:
            config_manager (ConfigManager): The ConfigManager instance to create the class instance from.

        Raises:
            ConfigError: If the provider is not specified or if the API key is not set.
        """
        self.config_manager = config_manager
        self.conversation_history: list[dict[str, str]] = []
        self.provider: str = self.config_manager.get(PROVIDER_KEY)
        if not self.provider:
            raise ConfigError(ConfigErrorEnum.PROVIDER_KEY_MISSING)

        if not self.config_manager.get(self.provider):
            raise ConfigError(ConfigErrorEnum.PROVIDER_CONFIG_MISSING, self.provider)

        self.api_key: str = self.config_manager.get(f"{self.provider}.api_key")
        if not self.config_manager.is_api_key_set:
            raise ConfigError(ConfigErrorEnum.API_KEY_MISSING, self.provider)

        self.model: str = self.config_manager.get(f"{self.provider}.model", DEFAULT_MODEL)
        self.api_base: str = self.config_manager.get(f"{self.provider}.api_base", DEFAULT_API_BASE)
        self.retries: int = int(
            self.config_manager.get(f"{self.provider}.retries", DEFAULT_RETRIES)
        )
        self.completion_path: str = self.config_manager.get(
            f"{self.provider}.completion_path", "/chat/completions"
        )
        self.content_path: str = self.config_manager.get(
            f"{self.provider}.answer_path", "choices.0.message.content"
        )
        self.proxy: str = self.config_manager.get(f"{self.provider}.proxy", "")
        self._request_timeout: int = int(self.config_manager.get(f"{self.provider}.timeout", 30))

        # Initialize HTTP clients with connection pooling
        limits = httpx.Limits(max_keepalive_connections=5, max_connections=10)
        transport = httpx.HTTPTransport(retries=self.retries, limits=limits)
        client_params = {"transport": transport, "timeout": self._request_timeout}

        if self.proxy:
            client_params["proxy"] = self.proxy
            logger.debug("Using proxy: %s", self.proxy)

        self._http_client = httpx.Client(**client_params)

        logger.debug(
            "Provider: %s, Model: %s, retries: %d", self.provider, self.model, self.retries
        )

    def generate(self, prompt: str, use_history: bool = False) -> str:
        """
        Generates a response based on the given prompt and conversation history.

        Args:
            prompt (str): The input prompt to generate a response for.
            use_history (bool, optional): Whether to use the conversation history. Defaults to False.

        Returns:
            str: The generated response.

        Raises:
            ConfigError: If the API key is not set in the config.
            BadRequestError: If the completion API returns an error.
        """
        console.print(f"Discovered model `{self.model}` with provider `{self.provider}`.")
        if use_history:
            messages = [*self.conversation_history, {"role": "user", "content": prompt}]
        else:
            messages = [{"role": "user", "content": prompt}]
        params: CompleteParams = self.gen_chat_params(messages)

        # Completion_with_retries returns a dictionary with the response and metadata
        # Could raise BadRequestError error
        console.print("🤖 Hang tight, I'm cooking up something good!")
        response: dict = self.completion_with_retries(**params)
        usage: dict = response.get("usage", {})

        assistant_message: str = glom(response, self.content_path, default="").strip()

        if use_history:
            self.conversation_history.append({"role": "user", "content": prompt})
            self.conversation_history.append({"role": "assistant", "content": assistant_message})
        if not usage:
            console.print("No usage response found.")
        else:
            text = Text("Token usage> prompt tokens: ")
            text.append(f"{usage.get('prompt_tokens')}", Colors.LIGHT_MAGENTA_RGB)
            text.append(", completion tokens: ")
            text.append(f"{usage.get('completion_tokens')}", Colors.LIGHT_MAGENTA_RGB)
            text.append(" total tokens: ")
            text.append(f"{usage.get('total_tokens')}", Colors.LIGHT_MAGENTA_RGB)
            console.print(text)

        return assistant_message

    def gen_chat_params(self, messages: t.Optional[list[dict]] = None) -> CompleteParams:
        """
        Generates the parameters for the chat completion API.

        Returns:
            CompleteParams: The parameters for the chat completion API.
        """
        params: CompleteParams = {
            "model": self.model,
            "api_key": self.api_key,
            "api_base": self.api_base,
            "messages": messages or [],
        }

        # set optional params
        max_tokens = int(self.config_manager.get(f"{self.provider}.max_tokens", 0))
        params["max_tokens"] = max_tokens or 100
        temperature = float(self.config_manager.get(f"{self.provider}.temperature", 0))
        if temperature:
            params["temperature"] = temperature
        top_p = float(self.config_manager.get(f"{self.provider}.top_p", 0.0))
        if top_p:
            params["top_p"] = top_p
        frequency_penalty = float(self.config_manager.get(f"{self.provider}.frequency_penalty", 0))
        if frequency_penalty:
            params["frequency_penalty"] = frequency_penalty
        try:
            extra_headers = json.loads(
                str(self.config_manager.get(f"{self.provider}.extra_headers", "{}"))
            )
        except json.JSONDecodeError:
            console.print(
                f"{self.provider}.extra_headers is not a valid JSON string, ignored.",
                style="yellow",
            )
            extra_headers = None
        if extra_headers:
            params["extra_headers"] = extra_headers
        return params

    def clear_history(self) -> None:
        """
        Clears the conversation history.
        """
        self.conversation_history = []

    def completion_with_retries(
        self,
        api_base,
        api_key,
        model,
        messages,
        max_tokens,
        temperature=None,
        top_p=None,
        frequency_penalty=None,
        extra_headers=None,
    ) -> dict:
        """
        Wrapper around the completion API that retries on failure.

        Args:
            api_base (str): The base URL for the API.
            api_key (str): The API key to use for authentication.
            model (str): The model to use for completion.
            messages (list[dict]): The messages to send to the API.
            max_tokens (int, optional): The maximum number of tokens to generate.
            temperature (float, optional): The temperature to use for completion.
            top_p (float, optional): The top_p to use for completion.
            frequency_penalty (float, optional): The frequency_penalty to use for completion.
            extra_headers (dict, optional): Additional headers to send with the request.

        Returns:
            dict: The response from the API.

        Raises:
            ConfigError: If the API key is not set in the config.
            BadRequestError: If the completion API returns an error.
        """
        headers = {"Authorization": f"Bearer {api_key}", "Content-Type": "application/json"}
        if extra_headers:
            headers.update(extra_headers)

        payload = {
            "model": model,
            "messages": messages,
            "max_tokens": max_tokens,
        }
        if temperature is not None:
            payload["temperature"] = temperature
        if top_p is not None:
            payload["top_p"] = top_p
        if frequency_penalty is not None:
            payload["frequency_penalty"] = frequency_penalty

        url = self._build_url(api_base)

        try:
            response = self._http_client.post(url, json=payload, headers=headers)
            response.raise_for_status()
            return response.json()
        except httpx.TimeoutException as e:
            logger.error("Request timed out: %s", str(e))
            raise
        except httpx.HTTPStatusError as e:
            logger.error("HTTP error occurred: %s", str(e))
            raise
        except Exception as e:
            logger.error("Unexpected error: %s", str(e))
            raise

    def _build_url(self, api_base: str) -> str:
        """Helper method to build the API URL."""
        if api_base.endswith("/"):
            api_base = api_base[:-1]
        if not self.completion_path.startswith("/"):
            self.completion_path = "/" + self.completion_path
        return f"{api_base}{self.completion_path}"

    def __del__(self):
        """Cleanup method to close HTTP clients."""
        if hasattr(self, "_http_client"):
            self._http_client.close()
