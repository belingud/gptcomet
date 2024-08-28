from typing import Annotated

import typer

from gptcomet.config_manager import ConfigManager, get_config_manager
from gptcomet.log import logger, set_debug


def entry(
    debug: Annotated[
        bool, typer.Option("--debug", "-d", help="Print debug information.")
    ] = False,
    local: Annotated[
        bool, typer.Option("--local", help="Use local configuration file.")
    ] = False,
):
    cfg: ConfigManager = get_config_manager(local=local)
    if debug:
        set_debug()
        logger.debug(f"Using Config path: {cfg.current_config_path}")
    cfg.reset()
    typer.echo(f"Configuration `{cfg.current_config_path}` reset to default values")
