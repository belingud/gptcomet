import logging
from pathlib import Path
from typing import Any, Optional, Union

import click
import tomlkit as toml
from tomlkit import TOMLDocument, boolean, float_, integer, string
from tomlkit.items import Item

from gptcomet._types import CacheType
from gptcomet.exceptions import KeyNotFound
from gptcomet.support_keys import SUPPORT_KEYS
from gptcomet.utils import is_float

logger = logging.getLogger(__name__)


class ConfigManager:
    __slots__ = (
        "local",
        "global_config_file",
        "default_config_file",
        "current_config_path",
        "_cache",
        "_valid_keys"
    )
    local: bool
    global_config_file: Path
    default_config_file: Path
    current_config_path: Path
    _cache: CacheType
    _valid_keys: set[str]

    def __init__(self, config_path: Path, local: bool = False):
        self.local: bool = local
        self.global_config_file: Path = Path.home() / ".local" / "gptcomet" / "gptcomet.toml"
        self.default_config_file: Path = Path(__file__).parent / "gptcomet.toml"
        # runtime config file
        self.current_config_path = config_path

        self._cache: CacheType = {"config": None, "default_config": None}
        self._valid_keys: set[str] = set(SUPPORT_KEYS.splitlines())

    @property
    def config(self):
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
    def default_config(self) -> toml.TOMLDocument:
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
        if local:
            return Path.cwd() / ".git" / "gptcomet.toml"
        else:
            return Path.home() / ".local" / "gptcomet" / "gptcomet.toml"

    def get_config_file(self, local: bool = False) -> Path:
        """
        Retrieves the path to the configuration file.

        If local is True, checks if the current directory is a git repository
        and returns the path to .git/gptcomet.toml if it is, otherwise returns
        the global configuration file path.

        Returns:
            Path: The path to the configuration file.
        """
        config_path = self.global_config_file
        if local:
            cwd = Path.cwd()
            if not (cwd / ".git").exists():
                click.echo(f"[{click.style('GPTComet', fg='yellow')}] Not a git repository. Using global config.")
            else:
                config_path = cwd / ".git" / "gptcomet.toml"
        # click.echo(f"[GPTComet] Using config file: {config_path}")
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
            click.echo(f"[{click.style('GPTComet', fg='green')}] Created default config file at {config_file}")

    def load_config(self) -> toml.TOMLDocument:
        """
        Load the configuration from the current configuration file.

        Returns:
            toml.TOMLDocument: The configuration loaded from the current configuration file.

        """
        self.ensure_config_file()
        with open(self.current_config_path) as f:
            return toml.load(f)

    def load_default_config(self) -> toml.TOMLDocument:
        """
        Load the default configuration from the default configuration file.

        Returns:
            toml.TOMLDocument: The default configuration loaded from the file.
        """
        with open(self.default_config_file) as f:
            return toml.load(f)

    def save_config(self):
        """
        Saves the current configuration to a file.
        Returns:
            None
        """
        with self.current_config_path.open("w") as f:
            f.write(self.config.as_string())

    def is_valid_key(self, key: str) -> bool:
        """
        Check if the given key is a valid key in the current configuration.

        Args:
            key (str): The key to be checked.

        Returns:
            bool: True if the key is valid, False otherwise.
        """
        return key in self._valid_keys

    def get_nested_value(self, doc: Union[dict, toml.TOMLDocument], keys: Union[str, list[str]]) -> Any:
        """
        Get the nested value from a dictionary or TOML document using a list of keys.

        Args:
            doc (Union[dict, toml.TOMLDocument]): The dictionary or TOML document to retrieve the nested value from.
            keys (Union[str, List[str]]): The list of keys to navigate through the document.

        Returns:
            Any: The nested value if found, None otherwise.
        """
        if isinstance(keys, str):
            keys = keys.split(".")
        for key in keys:
            if isinstance(doc, dict):
                doc = doc.get(key)
            else:
                raise KeyNotFound(key)
        return doc

    def set_nested_value(self, doc: TOMLDocument, keys: Union[str, list[str]], value: Any):
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
            >>> doc
            {'a': {'b': {'c': 2}}}

        Note:
            This function modifies the input dictionary `d` in-place.
        """
        if isinstance(keys, str):
            keys = keys.split(".")
        for key in keys[:-1]:
            doc = doc.setdefault(key, {})
        doc[keys[-1]] = value

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
        toml_value: Item = self.convert2toml_value(value)
        self.set_nested_value(self.config, key, toml_value)
        self.save_config()

    def get(self, key: str, default: Optional[Any] = None) -> Any:
        """
        Retrieves the value associated with the given key from the configuration.

        Args:
            key (str): The key to retrieve the value for.
            default (Optional[str]): The default value to return if the key is not found.

        Returns:
            Any: The value associated with the given key.

        Raises:
            KeyNotSupportError: If the key is not supported.
            ConfigKeyError: If the key does not exist in the configuration.

        Note:
            This method assumes that the `config` attribute of the `ConfigManager` object is a nested dictionary
            structure with keys separated by dots.
        """
        try:
            return self.get_nested_value(self.config, key)
        except KeyNotFound:
            return default

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
        with open(self.default_config_file) as default, open(self.current_config_path, "w") as target:
            content = default.read()
            target.write(content)
        self._cache["config"] = toml.loads(content)

    def list_keys(self) -> str:
        """
        Returns a list of supported keys.
        Returns:
            str: A list of supported keys.
        """
        return SUPPORT_KEYS

    def convert2toml_value(self, value: str) -> Item:
        """
        Converts a string value to a TOML-compatible value.

        Args:
            value (str): The string value to be converted.

        Returns:
            Item: The converted TOML value or the original string value.
        """
        if value.lower() in ("true", "false"):
            return boolean(value.lower())
        elif value.isnumeric():
            return integer(value)
        elif value.lower() in ("none", "null"):
            return string(value, literal=True)
        elif is_float(value):
            return float_(value)
        else:
            return string(value)
