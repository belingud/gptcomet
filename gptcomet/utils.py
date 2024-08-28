import fnmatch
import logging
import sys
import typing as t
from functools import wraps

import click
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
    if isinstance(value, bool):
        return value
    elif isinstance(value, int):
        return bool(value)
    true_values = {"ok", "true", "yes", "1", "y", "on"}
    false_values = {"false", "no", "0", "n", "off"}

    value = value.lower()
    if value in true_values:
        return True
    elif value in false_values:
        return False
    else:
        msg = f"{value} is not a valid boolean value, use {true_values | false_values}."
        raise ValueError(msg)


def common_options(func):
    """
    Decorator function that wraps another function with common options and saves them in the context object.

    Parameters:
        func (callable): The function to be wrapped.

    Example:
        >>> @click.command(help="My command")
        >>> @click.pass_context
        >>> @common_options
        >>> def my_command(ctx, debug, local):
        >>>     pass
    Returns:
        callable: The wrapped function.
    """

    @wraps(func)
    def wrapper(*args, **kwargs):
        ctx = (
            args[0] if args and isinstance(args[0], click.Context) else click.get_current_context()
        )
        ctx.ensure_object(dict)
        save_common_options(ctx)
        return func(*args, **kwargs)

    wrapper = click.option(
        "--local",
        is_flag=True,
        default=False,
        help="Use local configuration file or global.",
    )(wrapper)
    wrapper = click.option("--debug", is_flag=True, default=False, help="Enable debug mode")(
        wrapper
    )
    return wrapper


def save_common_options(ctx: click.Context):
    """Accept --debug and --local options in any command."""
    if ctx.obj.get("debug") is None or ctx.params["debug"] is True:
        ctx.obj["debug"] = ctx.params["debug"]
    if ctx.obj.get("local") is None or ctx.params["local"] is True:
        ctx.obj["local"] = ctx.params["local"]
    if ctx.obj["debug"] is True:
        logging.getLogger("gptcomet").setLevel(logging.DEBUG)


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
