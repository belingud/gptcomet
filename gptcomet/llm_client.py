import time
import typing as t

from gptcomet.exceptions import ConfigError, ConfigErrorEnum, GPTCometError
from gptcomet.llms import ProviderRegistry
from gptcomet.log import logger
from gptcomet.styles import Colors
from gptcomet.utils import console

if t.TYPE_CHECKING:
    from gptcomet.config_manager import ConfigManager


class LLMClient:
    """Client for interacting with LLM providers."""

    __slots__ = (
        "api_base",
        "api_key",
        "completion_path",
        "config_manager",
        "content_path",
        "conversation_history",
        "llm",
        "model",
        "provider_name",
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
        """Initialize the LLM client."""
        self.config_manager = config_manager
        self.conversation_history: list[dict[str, str]] = []

        # Get provider configuration
        self.provider_name = self.config_manager.get("provider")
        if not self.provider_name:
            raise ConfigError(ConfigErrorEnum.PROVIDER_KEY_MISSING)

        # Get provider configuration
        provider_config = ProviderRegistry.get_default_config(self.provider_name)
        user_config = self.config_manager.get(self.provider_name, {})
        provider_config.update(user_config)

        # Check API key
        if not provider_config.get("api_key"):
            raise ConfigError(ConfigErrorEnum.API_KEY_MISSING, self.provider_name)

        # Initialize provider
        provider_class = ProviderRegistry.get_provider(self.provider_name)
        self.llm = provider_class(provider_config)

        if self.config_manager.get("console.verbose"):
            console.print(
                f"Initialized {self.provider_name} provider with model {provider_config.get('model')}"
            )

        # Set configuration
        self.model: str = str(provider_config.get("model"))
        self.retries: int = int(provider_config.get("retries", 2))

        logger.debug(
            "Provider: %s, Model: %s, retries: %d", self.provider_name, self.model, self.retries
        )

    def generate(self, prompt: str, use_history: bool = False) -> str:
        """Generate text using the LLM provider."""
        if use_history:
            return self.completion_with_retries(prompt, history=self.conversation_history)
        return self.completion_with_retries(prompt)

    def completion_with_retries(
        self, message: str, history: t.Optional[list[dict[str, str]]] = None
    ) -> str:
        """Make a completion request with retries."""
        retries = self.retries
        while retries >= 0:
            try:
                result = self.llm.make_request(message, history)
                if result is None:
                    return ""  # Return an empty string if None is received
                else:
                    return result
            except Exception as e:
                console.print(
                    f"Request failed with error: {e}. Retrying in 1 second... ({retries} retries left)",
                    style=Colors.YELLOW,
                )
                if retries == 0:
                    raise e from None
                retries -= 1
                time.sleep(1)
        msg = f"Request failed after {self.retries} retries"
        raise GPTCometError(msg)

    def clear_history(self) -> None:
        """
        Clears the conversation history.
        """
        self.conversation_history = []
