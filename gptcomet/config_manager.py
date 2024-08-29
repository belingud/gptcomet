from pathlib import Path
from typing import Any, Optional, Union

import click
import orjson as json
from glom import assign, glom
from ruamel.yaml import YAML, CommentedMap

from gptcomet._types import CacheType
from gptcomet.const import LANGUAGE_KEY
from gptcomet.exceptions import (
    ConfigKeyError,
    ConfigKeyTypeError,
    LanguageNotSupportError,
    NotModified,
)
from gptcomet.support_keys import SUPPORT_KEYS
from gptcomet.utils import convert2type, output_language_map, strtobool

yaml = YAML(typ="rt", pure=True)


class ConfigManager:
    __slots__ = (
        "current_config_path",
        "_cache",
        "_valid_keys",
    )
    current_config_path: Path
    _cache: CacheType

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
    def get_config_path(cls, local: bool = False) -> Path:
        return Path.cwd() / ".git" / "gptcomet.yaml" if local else cls.global_config_file

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
            doc (TOMLDocument): The dictionary to modify.
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
        if isinstance(keys, (list, tuple, set)):
            keys = ".".join(keys)
        assign(doc, keys, value)

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
        with self.current_config_path.open() as f:
            return f.read()

    def reset(self):
        """
        Resets the current configuration to its default state.
        """
        with (
            self.default_config_file.open() as default,
            self.current_config_path.open("w") as target,
        ):
            content = default.read()
            target.write(content)
            self._cache["config"] = yaml.load(default)

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
        try:
            return strtobool(value)
        except ValueError:
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
    return ConfigManager.from_config_path(ConfigManager.get_config_path(local=local))
