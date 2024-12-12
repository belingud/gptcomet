from typing import Annotated, Union

import orjson
import typer
from ruamel.yaml.comments import CommentedMap, CommentedSeq

from gptcomet.config_manager import ConfigManager, get_config_manager
from gptcomet.exceptions import ConfigKeyError
from gptcomet.log import logger, set_debug
from gptcomet.utils import mask_api_keys


def entry(
    key: Annotated[str, typer.Argument(..., help="The configuration key to get the value for.")],
    debug: Annotated[bool, typer.Option("--debug", "-d", help="Print debug information.")] = False,
    local: Annotated[bool, typer.Option("--local", help="Use local configuration file.")] = False,
):
    cfg: ConfigManager = get_config_manager(local=local)
    if debug:
        set_debug()
        logger.debug(f"Using Config path: {cfg.current_config_path}")
    try:
        value: Union[str, CommentedSeq, CommentedMap] = cfg.get(key)
        value = mask_api_keys(value)
        styled_key: str = typer.style(key, fg=typer.colors.GREEN)
        v = orjson.dumps(value, option=orjson.OPT_INDENT_2).decode("utf-8")
        typer.echo(f"{styled_key}: {typer.style(v, fg=typer.colors.GREEN)}")
    except ValueError as e:
        typer.echo(f"Error: {e!s}")
    except ConfigKeyError as e:
        typer.echo(str(e))
