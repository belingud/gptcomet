import fnmatch
import sys
import typing as t

from rich import get_console

defenc = sys.getdefaultencoding()
console = get_console()

# context
CONTEXT_SETTINGS = {"help_option_names": ["-h", "--help"]}

LIST_VALUES = ("file_ignore",)

output_language_map = {
    "en": "English",
    "zh-cn": "Simplified Chinese",
    "zh-tw": "Traditional Chinese",
    "fr": "French",
    "vi": "Vietnamese",
    "ja": "Japanese",
    "ko": "Korean",
    "ru": "Russian",
    "tr": "Turkish",
    "id": "Indonesian",
    "th": "Thai",
    "de": "German",
    "es": "Spanish",
}


def convert2type(v: t.Any, _type: t.Optional[type]) -> t.Any:
    if _type is None:
        return v
    if not callable(_type):
        raise TypeError(f"_type must be callable, got {_type}.")  # noqa: TRY003
    try:
        return _type(v)
    except ValueError:
        msg = f"Could not convert value {v} to type {_type}."
        raise ValueError(msg) from None


def strtobool(value: t.Union[str, bool, int]) -> bool:
    """
    Convert a string to a boolean value.

    Parameters:
        value (Union[str, bool, int]): The string to convert.

    Returns:
        bool: The boolean value of the string.

    Raises:
        ValueError: If the string cannot be converted to a boolean value.
    """
    true_values = {"ok", "true", "yes", "1", "y", "on"}
    false_values = {"false", "no", "0", "n", "off"}
    if not isinstance(value, (str, bool, int)):
        msg = f"{value} is not a valid boolean value, use {true_values | false_values}."
        raise TypeError(msg)
    if isinstance(value, bool):
        return value
    elif isinstance(value, int):
        return bool(value)

    value = value.lower()
    if value in true_values:
        return True
    elif value in false_values:
        return False
    else:
        msg = f"{value} is not a valid boolean value, use {true_values | false_values}."
        raise ValueError(msg)


def is_float(s):
    """
    Checks if a given string can be converted to a float.

    Args:
        s (str): The string to check.

    Returns:
        bool: True if the string can be converted to a float, False otherwise.
    """
    try:
        float(s)
    except ValueError:
        return False
    else:
        return True


def should_ignore(filepath: str, ignore_patterns: list[str]) -> bool:
    """
    Checks if a given filepath should be ignored based on the ignore patterns.

    Args:
        filepath (str): The filepath to check.
        ignore_patterns (list[str]): The list of ignore patterns.

    Returns:
        bool: True if the filepath should be ignored, False otherwise.
    """
    return any(fnmatch.fnmatch(filepath, pattern) for pattern in ignore_patterns)
