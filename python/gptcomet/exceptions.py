import enum
import typing as t

from gptcomet.utils import output_language_map


class GPTCometError(Exception):
    pass


class NotModified(GPTCometError):
    REASON_EXISTS = "Config value already exists and not modified"
    REASON_EMPTY = "Config value is empty"


class KeyNotFound(GPTCometError):
    def __init__(self, key: str):
        self.key = key

    def __str__(self):
        return f"Key '{self.key}' not found in the configuration."


class GitNoStagedChanges(GPTCometError):
    def __str__(self):
        return "No staged changes to commit"


class ConfigErrorEnum(enum.IntEnum):
    """Enum for config error."""

    def __new__(cls, code, description):
        obj = int.__new__(cls, code)
        obj._value_ = code
        obj.description = description
        return obj

    API_KEY_MISSING = 0, "Missing API key in config file for provider '{provider}'"
    PROVIDER_KEY_MISSING = 1, "No LLM provider specified in config file"
    PROVIDER_CONFIG_MISSING = 2, "Configuration for provider '{provider}' not found"


class ConfigError(GPTCometError):
    def __init__(
        self,
        error: ConfigErrorEnum = ConfigErrorEnum.API_KEY_MISSING,
        provider: t.Optional[str] = None,
    ):
        if not isinstance(error, ConfigErrorEnum):
            raise TypeError
        self.error = error
        self.provider = provider or "unknown"

    def __str__(self):
        return self.error.description.format(provider=self.provider)


class ConfigKeyError(GPTCometError):
    def __init__(self, key: str):
        self.key = key

    def __str__(self):
        return (
            f"Key '{self.key}' is not allowed. "
            f"Only specified keys can be set, use 'gptcomet config keys' to see the allowed keys."
        )


class ConfigKeyTypeError(GPTCometError):
    def __init__(self, key: str, need_type: str = "list"):
        self.key = key
        self.need_type = need_type

    def __str__(self):
        return f"Key '{self.key}' is not a {self.need_type} type."


class LanguageNotSupportError(GPTCometError):
    def __init__(self, lang: str):
        self.lang = lang
        self.supported = list(output_language_map.keys())

    def __str__(self):
        return f"Language '{self.lang}' not support. Choose from {self.supported}"


class KeyNotSupportError(GPTCometError):
    def __init__(self, key: str):
        self.key = key

    def __str__(self):
        return f"Key '{self.key}' not support."
