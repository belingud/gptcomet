from typing import Annotated

import typer

from gptcomet.config_manager import ConfigManager, get_config_manager
from gptcomet.utils import console


def entry(
    local: Annotated[
        bool,
        typer.Option("--local", help="Use local configuration file.", rich_help_panel="Options"),
    ] = False,
):
    cfg: ConfigManager = get_config_manager(local=local)
    console.print(cfg.current_config_path)
