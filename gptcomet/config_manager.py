import json
from dataclasses import dataclass, field
from io import StringIO
from pathlib import Path
from typing import Any, Optional, Union

import click
from glom import glom
from ruamel.yaml import YAML, CommentedMap

from gptcomet._types import CacheType
from gptcomet.const import LANGUAGE_KEY, PROVIDER_KEY
from gptcomet.exceptions import (
    ConfigKeyError,
    ConfigKeyTypeError,
    LanguageNotSupportError,
    NoSuchProvider,
    NotModified,
)
from gptcomet.support_keys import SUPPORT_KEYS
from gptcomet.utils import convert2type, mask_api_keys, output_language_map, strtobool

yaml = YAML(typ="rt", pure=True)


@dataclass
class ProviderConfig:
    """Provider configuration data class."""

    provider: str = field(default="openai")
    api_base: str = field(default="https://api.openai.com/v1/")
    model: str = field(default="")
    api_key: str = field(default="")
    max_tokens: int = field(default=1024)
    retries: int = field(default=2)
    _extra_fields: dict[str, Any] = field(default_factory=dict, repr=False)

    def __init__(
        self,
        provider: str = "openai",
        api_base: str = "https://api.openai.com/v1/",
        model: str = "",
        api_key: str = "",
        max_tokens: int = 1024,
        retries: int = 2,
        **kwargs: Any,
    ) -> None:
        """Initialize provider config with optional extra fields."""
        self.provider = provider
        self.api_base = api_base
        self.model = model
        self.api_key = api_key
        self.max_tokens = max_tokens
        self.retries = retries
        self._extra_fields = kwargs

    def to_dict(self) -> dict[str, Any]:
        """Convert to provider dictionary."""
        base_config = {
            "api_base": self.api_base,
            "api_key": self.api_key,
            "model": self.model,
            "max_tokens": self.max_tokens,
            "retries": self.retries,
        }
        return {**base_config, **self._extra_fields}


class ConfigManager:
    __slots__ = (
        "_cache",
        "_valid_keys",
        "current_config_path",
    )
    current_config_path: Path

    global_config_file: Path = Path.home() / ".config" / "gptcomet" / "gptcomet.yaml"
    default_config_file: Path = Path(__file__).parent / "gptcomet.yaml"

    def __init__(self, config_path: Path):
        # runtime config file
        self.current_config_path = config_path

        self._cache: CacheType = {"config": None, "default_config": None}

    @classmethod
    def from_config_path(cls, config_path: Path) -> "ConfigManager":
        return cls(config_path)

    @property
    def config(self) -> CommentedMap:
        """
        Returns the configuration stored in the cache.

        If the configuration is not already loaded, it is loaded by calling the
        `load_config` method and stored in the cache.

        Returns:
            The configuration stored in the cache.
        """
        if "config" not in self._cache or self._cache["config"] is None:
            self._cache["config"] = self.load_config()
        return self._cache["config"]

    @property
    def default_config(self) -> CommentedMap:
        """
        Returns the default configuration stored in the cache.

        If the default configuration is not already loaded, it is loaded by calling the
        `load_default_config` method and stored in the cache.

        Returns:
            The default configuration stored in the cache.
        """
        if "default_config" not in self._cache or self._cache["default_config"] is None:
            self._cache["default_config"] = self.load_default_config()
        return self._cache["default_config"]

    @classmethod
    def make_config_path(cls, local: bool = False) -> Path:
        return Path.cwd() / ".git" / "gptcomet.yaml" if local else cls.global_config_file

    def set_cli_overrides(
        self, provider: Optional[str] = None, api_config: Optional[dict[str, Any]] = None
    ) -> None:
        """
        Set command line overrides for configuration values.

        Args:
            provider (Optional[str]): The provider to use.
            api_config (Dict[str, Any]): The API configuration to use.
        """
        if not api_config:
            api_config = {}
        overrides = {}
        if provider:
            if not isinstance(provider, str):
                msg = "Provider must be a string"
                raise TypeError(msg)
            overrides["provider"] = provider
        api_config = {k: v for k, v in api_config.items() if v is not None}
        if api_config:
            provider_from_file = self.get("provider")
            provider_from_cli = provider or provider_from_file
            if not provider_from_cli:
                msg = "Provider is required when using --api-base or --api-key"
                raise ValueError(msg)
            existing_config = self.get(provider_from_cli, {})
            if isinstance(existing_config, dict):
                merged_api_config = {**existing_config, **api_config}
            else:
                merged_api_config = api_config
            overrides[provider_from_cli] = merged_api_config
        self._cache["config"] = {**self.config, **overrides}

    def is_api_key_set(self) -> bool:
        """
        Checks if the API key is set in the config file.

        Returns:
            bool: True if the API key is set, False if is empty or default.
        """
        if not self.current_config_path.exists():
            return False
        provider = self.get("provider")
        api_key: str = self.get(f"{provider}.api_key")
        return str(api_key).strip("x") != "sk-"

    def get_config_file(self, local: bool = False) -> Path:
        """
        Retrieves the path to the configuration file.

        If local is True, checks if the current directory is a git repository
        and returns the path to .git/gptcomet.yaml if it is, otherwise returns
        the global configuration file path.

        Returns:
            Path: The path to the configuration file.
        """
        config_path = self.global_config_file
        if local:
            cwd = Path.cwd()
            if not (cwd / ".git").exists():
                click.echo(
                    f"[{click.style('GPTComet', fg='yellow')}] Not a git repository. Using global config."
                )
            else:
                config_path = cwd / ".git" / "gptcomet.yaml"
        # click.echo(f"{GPTCOMET_PRE} Using config file: {config_path}")
        return config_path

    def ensure_config_file(self):
        """
        Ensures that the configuration file exists by creating it with default values if it does not already exist.
        """
        config_file = self.current_config_path
        if not config_file.exists():
            config_file.parent.mkdir(parents=True, exist_ok=True)
            with open(self.default_config_file) as default, open(config_file, "w") as target:
                target.write(default.read())
            click.echo(
                f"[{click.style('GPTComet', fg='green')}] Created default config file at {config_file}"
            )

    def load_config(self) -> CommentedMap:
        """
        Load the configuration from the current configuration file.

        Returns:
            CommentedMap: The configuration loaded from the current configuration file.

        """
        self.ensure_config_file()
        with self.current_config_path.open() as f:
            return yaml.load(f)

    def load_default_config(self) -> CommentedMap:
        """
        Load the default configuration from the default configuration file.

        Returns:
            CommentedMap: The default configuration loaded from the file.
        """
        with self.default_config_file.open() as f:
            return yaml.load(f)

    def save_config(self):
        """
        Saves the current configuration to a file.
        """
        with self.current_config_path.open("w") as f:
            yaml.dump(self.config, f)

    def get_nested_value(
        self, doc: Union[dict, CommentedMap], keys: Union[str, list[str]], default: Any = None
    ) -> Any:
        """
        Get the nested value from a dictionary or TOML document using a list of keys.

        Args:
            doc (Union[dict, CommentedMap]): The dictionary or TOML document to retrieve the nested value from.
            keys (Union[str, List[str]]): The list of keys to navigate through the document.
            default (Any, optional): The default value to return if the nested value is not found. Defaults to None.

        Returns:
            Any: The nested value if found, None otherwise.
        """
        if isinstance(keys, list):
            keys = ".".join(keys)
        return glom(doc, keys, default=default)

    def set_nested_value(self, doc: CommentedMap, keys: Union[str, list[str]], value: Any):
        """
        Set a nested value in a dictionary using a list of keys.

        Args:
            doc (CommentedMap): The dictionary to modify.
            keys (Union[str, List[str]]): The list of keys to navigate through the dictionary.
            value (Any): The value to set at the end of the nested keys.

        It navigates through the dictionary using the keys and sets the value at the end of the nested keys.
        If any key in the path does not exist, it creates a new dictionary at that key.

        Example:
            >>> config = ConfigManager()
            >>> d = {'a': {'b': {'c': 1}}}
            >>> config.set_nested_value(doc, ['a', 'b', 'c'], 2)
            >>> d
            {'a': {'b': {'c': 2}}}

        Note:
            This function modifies the input dictionary `d` in-place.
        """
        if isinstance(keys, str):
            keys = keys.split(".")
        for key in keys[:-1]:
            doc = doc.setdefault(key, {})
        doc[keys[-1]] = value

    def add_provider(self, provider: str):
        if provider in self.config:
            raise ValueError(f"Provider {provider} already exists in config.")  # noqa: TRY003
        info = {
            provider: {
                "model": "xxx",
                "api_key": "xxx",
                "proxy": "",
                "max_tokens": 4096,
                "temperature": 0.7,
                "top_p": 1,
                "frequency_penalty": 0,
                "presence_penalty": 0,
            }
        }
        self.config.update(info)

    def set_new_provider(self, provider: str, provider_info: dict):
        self.config[provider] = provider_info
        self.config["provider"] = provider
        self.save_config()

    def set(self, key: str, value: str):
        """
        Set the value of a configuration key.

        Args:
            key (str): The key to set the value for.
            value (str): The value to set.

        Raises:
            ConfigKeyError: If the key is not supported.

        This method sets the value of a configuration key in the `config` attribute of the `ConfigManager` object.
        It first checks if the key is valid using the `is_valid_key` method. If the key is not valid, it raises a
        `ConfigKeyError` with the invalid key.

        Note:
            This method modifies the toml `config` attribute in-place.
        """
        if key == LANGUAGE_KEY and output_language_map.get(value) is None:
            raise LanguageNotSupportError(key)
        if key.split(".")[-1] not in SUPPORT_KEYS:
            raise ConfigKeyError(key)
        if key == PROVIDER_KEY and not self.config.get(value):
            raise NoSuchProvider(value)
        toml_value = self.convert2yaml_value(value)
        self.set_nested_value(self.config, key, toml_value)
        self.save_config()

    def get(self, key: str, default: Optional[Any] = None, _type: Optional[type] = None) -> Any:
        """
        Retrieves the value associated with the given key from the configuration.

        Args:
            key (str): The key to retrieve the value for.
            default (Optional[str]): The default value to return if the key is not found.
            _type (Optional[Type]): The type to convert the retrieved value to.

        Returns:
            Any: The value associated with the given key.

        Raises:
            KeyNotSupportError: If the key is not supported.
            ConfigKeyError: If the key does not exist in the configuration.

        Note:
            This method assumes that the `config` attribute of the `ConfigManager` object is a nested dictionary
            structure with keys separated by dots.
        """
        v = self.get_nested_value(self.config, key, default=default)
        return convert2type(v=v, _type=_type)

    def list(self) -> str:
        """
        Reads the contents of the current configuration file and returns them as a string.

        Returns:
            str: The contents of the current configuration file.

        Raises:
            FileNotFoundError: If the current configuration file does not exist.
        Note:
            This method assumes that `self.current_config_path` is a valid file path.
        """

        config_data = self.config.copy()
        config_data.pop("prompt", None)

        # recursively mask api keys
        mask_api_keys(config_data, show_first=3)

        # convert to YAML string
        output = StringIO()
        yaml.dump(config_data, output)
        return output.getvalue()

    def reset(self, prompt: bool = False):
        """
        Resets the current configuration to its default state.

        Args:
            prompt: If True, only reset prompt configuration
        """
        default_config = self.load_default_config()
        if prompt:
            self.set("prompt", default_config.get("prompt"))
            return
        self._cache["config"] = default_config
        self.save_config()
        self._cache["config"] = None

    def list_keys(self) -> str:
        """
        Returns a list of supported keys.
        Returns:
            str: A list of supported keys.
        """
        return SUPPORT_KEYS

    def append(self, key: str, value: str):
        """
        Append a value to a configuration key.

        Args:
            key (str): The key to append the value to.
            value (str): The value to append.

        Raises:
            ConfigKeyError: If the key is not supported.

        This method appends a value to a configuration key in the `config` attribute of the `ConfigManager` object.
        It first checks if the key is valid using the `is_valid_key` method. If the key is not valid, it raises a
        `ConfigKeyError` with the invalid key.

        Note:
            This method modifies the toml `config` attribute in-place.
        """
        current_value = self.get(key)
        if not isinstance(current_value, list):
            raise ConfigKeyTypeError(key, "list")
        if value in current_value:
            raise NotModified(NotModified.REASON_EXISTS)
        current_value.append(value)
        self.save_config()

    def remove(self, key: str, value: str):
        """
        Remove a configuration key.

        Args:
            key (str): The key containing the value to remove.
            value (str): The value to remove.

        Raises:
            ConfigKeyError: If the key is not supported.

        This method removes a configuration key in the `config` attribute of the `ConfigManager` object.
        It first checks if the key is valid. If the key is not valid, it raises a
        `ConfigKeyTypeError` with the invalid key.

        Raises:
            NotModified: If the value is not found in the list.
            ConfigKeyTypeError: If the value is not a list.
            ValueError: If the value not found.

        Note:
            This method modifies the toml `config` attribute in-place.
        """
        current_value = self.get(key)
        if not isinstance(current_value, list):
            raise ConfigKeyTypeError(key)
        if not current_value:
            raise NotModified(NotModified.REASON_EMPTY)
        current_value.remove(value)
        self.save_config()

    def convert2yaml_value(self, value: str) -> Any:
        """
        Converts a string value to a YAML-compatible value.

        Args:
            value (str): The string value to be converted.

        Returns:
            Any: The converted YAML value or the original string value.
        """
        if not isinstance(value, str):
            return value
        try:
            return strtobool(value)
        except (ValueError, TypeError):
            pass
        if value.lower() in ("none", "null"):
            return None
        if value.isnumeric():
            return int(value)
        try:
            return float(value)
        except ValueError:
            pass
        try:
            # try to parse the value as a literal expression
            return json.loads(value)
        except json.JSONDecodeError:
            pass
        return value


def get_config_manager(local: bool) -> ConfigManager:
    """
    Load the configuration from the current configuration file.

    Args:
        local (bool): Whether to use the local configuration file.

    Returns:
        ConfigManager: The configuration manager with the configuration loaded from the current configuration file.
    """
    return ConfigManager.from_config_path(ConfigManager.make_config_path(local=local))
