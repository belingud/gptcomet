from pathlib import Path
from typing import Annotated, Optional

import typer

from gptcomet._types import Provider
from gptcomet.config_manager import ConfigManager, get_config_manager
from gptcomet.const import GPTCOMET_PRE
from gptcomet.log import logger, set_debug
from gptcomet.styles import Colors, stylize
from gptcomet.utils import console, raw_input


def entry(
    debug: Annotated[bool, typer.Option("--debug", "-d", help="Enable debug mode.")] = False,
    local: Annotated[bool, typer.Option("--local", "-l", help="Enable local mode.")] = False,
    config_path: Annotated[
        Optional[Path],
        typer.Option(
            "--config",
            "-c",
            help="Path to config file.",
            exists=True,
            file_okay=True,
            dir_okay=False,
            writable=False,
            readable=True,
            resolve_path=True,
        ),
    ] = None,
):
    """Setup new provider."""
    cfg: ConfigManager = (
        ConfigManager(config_path=config_path) if config_path else get_config_manager(local=local)
    )
    if debug:
        set_debug()
        logger.debug(f"Using config file: {cfg.current_config_path}")
    provider: str = typer.prompt("Enter provider(lowercase)", default="openai", type=str).lower()
    api_base: str = typer.prompt("Enter API Base", default="https://api.openai.com/v1/")
    model: str = typer.prompt("Enter model", default="text-davinci-003", type=str)

    console.print("Enter API key: ", end="")
    api_key: str = raw_input("Enter API key", mask=True)

    max_tokens: int = typer.prompt("Max tokens", default=1024, type=int)
    provider_cfg: Provider = {
        "api_base": api_base,
        "api_key": api_key,
        "model": model,
        "max_tokens": max_tokens,
        "retries": 2,
    }
    cfg.set_new_provider(provider, provider_cfg)
    console.print(stylize(f"{GPTCOMET_PRE} Provider {provider} set.", Colors.GREEN))
