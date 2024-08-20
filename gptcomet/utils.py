import logging
from functools import wraps

import click

logger = logging.getLogger(__name__)


LIST_VALUES = (
    "file_ignore",
)


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
        ctx = args[0] if args and isinstance(args[0], click.Context) else click.get_current_context()
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
