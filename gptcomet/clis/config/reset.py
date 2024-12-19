from typing import Annotated

import typer

from gptcomet.config_manager import ConfigManager, get_config_manager
from gptcomet.log import set_debug
from gptcomet.utils import console


def entry(
    prompt: Annotated[
        bool,
        typer.Option(
            "--prompt", help="Reset prompt to default of current version.", rich_help_panel="Prompt"
        ),
    ] = False,
    debug: Annotated[bool, typer.Option("--debug", "-d", help="Print debug information.")] = False,
    local: Annotated[bool, typer.Option("--local", help="Use local configuration file.")] = False,
):
    cfg: ConfigManager = get_config_manager(local=local)
    if debug:
        set_debug()
    console.print(f"Using Config path: {cfg.current_config_path}")
    cfg.reset(prompt=prompt)
    typer.echo(f"Configuration `{cfg.current_config_path}` reset to default values")
