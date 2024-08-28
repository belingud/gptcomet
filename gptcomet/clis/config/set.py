from typing import Annotated

import typer

from gptcomet.config_manager import ConfigManager, get_config_manager
from gptcomet.const import GPTCOMET_PRE
from gptcomet.exceptions import ConfigKeyError
from gptcomet.log import logger, set_debug


def entry(
    key: Annotated[str, typer.Argument(..., help="Configuration key to set.")],
    value: Annotated[str, typer.Argument(..., help="Value to set the configuration key to.")],
    debug: Annotated[bool, typer.Option("--debug", "-d", help="Print debug information.")] = False,
    local: Annotated[bool, typer.Option("--local", help="Use local configuration file.")] = False,
):
    cfg: ConfigManager = get_config_manager(local=local)
    if debug:
        set_debug()
        logger.debug(f"Using Config path: {cfg.current_config_path}")
    try:
        cfg.set(key, value)
        styled_key: str = typer.style(key, fg=typer.colors.GREEN)
        styled_value: str = typer.style(value, fg=typer.colors.GREEN)
        typer.echo(f"{GPTCOMET_PRE} Set {styled_key} to {styled_value}.")
    except ConfigKeyError as e:
        typer.echo(f"{GPTCOMET_PRE} Error: {e!s}")
