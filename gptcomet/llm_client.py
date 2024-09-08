import typing as t

import click
import orjson as json

try:
    import socksio  # noqa: F401
except ImportError as e:
    msg = e.msg
    if "socksio" in msg:
        msg = ("Using SOCKS proxy, but the 'socksio' package is not installed. "
               "Make sure to install gptcomet using `pip install gptcomet[socks]` or `pip install socksio`.")
    raise ImportError(msg) from None
from litellm import completion_with_retries
from litellm.types.utils import ModelResponse

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
        "config_manager",
        "conversation_history",
        "provider",
        "api_key",
        "model",
        "api_base",
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
        self.config_manager = config_manager
        self.conversation_history: list[dict[str, str]] = []
        self.provider: str = self.config_manager.get(PROVIDER_KEY)
        self.api_key: str = self.config_manager.get(f"{self.provider}.api_key")
        if not self.config_manager.is_api_key_set:
            raise ConfigError(ConfigErrorEnum.API_KEY_MISSING)
        self.model: str = self.config_manager.get(f"{self.provider}.model", DEFAULT_MODEL)
        self.api_base: str = self.config_manager.get(f"{self.provider}.api_base", DEFAULT_API_BASE)
        self.retries: int = int(
            self.config_manager.get(f"{self.provider}.retries", DEFAULT_RETRIES)
        )
        logger.debug(f"Provider: {self.provider}, Model: {self.model}, retries: {self.retries}")

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
        if use_history:
            messages = [*self.conversation_history, {"role": "user", "content": prompt}]
        else:
            messages = [{"role": "user", "content": prompt}]
        params = self.gen_chat_params(messages)

        # Completion_with_retries returns a dictionary with the response and metadata
        # Could raise BadRequestError error
        response: ModelResponse = completion_with_retries(**params)
        logger.debug(f"Response: {response}")

        assistant_message: str = response["choices"][0]["message"]["content"].strip()

        if use_history:
            self.conversation_history.append({"role": "user", "content": prompt})
            self.conversation_history.append({"role": "assistant", "content": assistant_message})

        return assistant_message

    def gen_chat_params(self, messages: t.Optional[list[dict]] = None) -> CompleteParams:
        """
        Generates the parameters for the chat completion API.

        Returns:
            CompleteParams: The parameters for the chat completion API.
        """
        params: CompleteParams = {
            "model": f"{self.provider}/{self.model}",
            "api_key": self.api_key,
            "api_base": self.api_base,
            "messages": messages or [],
        }

        # set optional params
        max_tokens = int(self.config_manager.get(f"{self.provider}.max_tokens"))
        if max_tokens:
            params["max_tokens"] = max_tokens
        temperature = float(self.config_manager.get(f"{self.provider}.temperature"))
        if temperature:
            params["temperature"] = temperature
        top_p = float(self.config_manager.get(f"{self.provider}.top_p"))
        if top_p:
            params["top_p"] = top_p
        frequency_penalty = float(self.config_manager.get(f"{self.provider}.frequency_penalty"))
        if frequency_penalty:
            params["frequency_penalty"] = frequency_penalty
        try:
            extra_headers = json.loads(
                str(self.config_manager.get(f"{self.provider}.extra_headers", "{}"))
            )
        except json.JSONDecodeError:
            click.echo(
                click.style(
                    f"{self.provider}.extra_headers is not a valid JSON string, ignored.",
                    fg="yellow",
                )
            )
            extra_headers = None
        if extra_headers:
            params["extra_headers"] = extra_headers
        return params

    def clear_history(self):
        """
        Clears the conversation history.
        """
        self.conversation_history = []
