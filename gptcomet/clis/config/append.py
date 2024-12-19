from typing import Annotated

import typer

from gptcomet.config_manager import ConfigManager, get_config_manager
from gptcomet.const import GPTCOMET_PRE
from gptcomet.exceptions import ConfigKeyTypeError, NotModified
from gptcomet.log import set_debug
from gptcomet.styles import Colors, stylize
from gptcomet.utils import console


def entry(
    key: Annotated[str, typer.Argument(..., help="Configuration key to set.")],
    value: Annotated[str, typer.Argument(..., help="Value to set the configuration key to.")],
    debug: Annotated[bool, typer.Option("--debug", "-d", help="Print debug information.")] = False,
    local: Annotated[bool, typer.Option("--local", help="Use local configuration file.")] = False,
):
    """Append a value to the list set by the corresponding key."""
    cfg: ConfigManager = get_config_manager(local=local)
    if debug:
        set_debug()
    console.print(f"Using Config path: {cfg.current_config_path}")
    try:
        cfg.append(key, value)
        console.print(stylize(f"{GPTCOMET_PRE} Appended {value} to {key}.", Colors.GREEN))

    except NotModified:
        console.print(
            stylize(
                f"{GPTCOMET_PRE} Config value already exists and not modified: {key!s}",
                Colors.LIGHT_BLUE_RGB,
            )
        )
    except ConfigKeyTypeError as e:
        console.print(stylize(f"{GPTCOMET_PRE} Error: {e!s}", Colors.LIGHT_RED_RGB))
