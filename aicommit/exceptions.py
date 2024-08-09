import enum


class AICommitError(Exception):
    pass


class KeyNotFound(AICommitError):
    def __init__(self, key: str):
        self.key = key

    def __str__(self):
        return f"Key '{self.key}' not found in the configuration."


class GitNoStagedChanges(AICommitError):

    def __str__(self):
        return "No staged changes to commit"


class ConfigErrorEnum(enum.IntEnum):
    """Enum for config error."""

    def __new__(cls, code, description):
        obj = int.__new__(cls, code)
        obj._value_ = code
        obj.description = description
        return obj

    API_KEY_MISSING = 0, "Missing {provider}.api_key in config file"


class ConfigError(AICommitError):

    def __init__(self, error: ConfigErrorEnum = ConfigErrorEnum.API_KEY_MISSING):
        if not isinstance(error, ConfigErrorEnum):
            raise TypeError
        self.error = error

    def __str__(self):
        return f"Config error: {self.error.description}"


class ConfigKeyError(AICommitError):
    def __init__(self, key: str):
        self.key = key

    def __str__(self):
        return (
            f"Key '{self.key}' is not allowed. "
            f"Only specified keys can be set, use 'aicommit config keys' to see the allowed keys."
        )


class KeyNotSupportError(AICommitError):
    def __init__(self, key: str):
        self.key = key

    def __str__(self):
        return f"Key '{self.key}' not support."
