import json
from typing import Annotated, Any

import typer

from gptcomet.config_manager import ConfigManager, get_config_manager
from gptcomet.exceptions import ConfigKeyError
from gptcomet.log import set_debug
from gptcomet.utils import console, mask_api_keys


def entry(
    key: Annotated[str, typer.Argument(..., help="The configuration key to get the value for.")],
    debug: Annotated[bool, typer.Option("--debug", "-d", help="Print debug information.")] = False,
    local: Annotated[bool, typer.Option("--local", help="Use local configuration file.")] = False,
) -> None:
    cfg: ConfigManager = get_config_manager(local=local)
    if debug:
        set_debug()
    console.print(f"Using Config path: {cfg.current_config_path}")
    try:
        value: Any = cfg.get(key)
        if isinstance(value, (dict, list)) or "api_key" in key:
            value = mask_api_keys(value)
        styled_key: str = typer.style(key, fg=typer.colors.GREEN)
        v: str = json.dumps(value, indent=2)
        typer.echo(f"{styled_key}: {typer.style(v, fg=typer.colors.GREEN)}")
    except ValueError as e:
        typer.echo(f"Error: {e!s}")
    except ConfigKeyError as e:
        typer.echo(str(e))
