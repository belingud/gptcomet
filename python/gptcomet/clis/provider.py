from pathlib import Path
from typing import Annotated, Optional

import typer
from prompt_toolkit import prompt

from gptcomet.config_manager import ConfigManager, ProviderConfig, get_config_manager
from gptcomet.const import GPTCOMET_PRE
from gptcomet.exceptions import ConfigError
from gptcomet.log import logger, set_debug
from gptcomet.styles import Colors, stylize
from gptcomet.utils import console


def create_provider_config() -> ProviderConfig:
    """Create provider config from user input."""
    try:
        provider = typer.prompt(
            "Enter provider name (lowercase)", default="openai", type=str
        ).lower()

        api_base = typer.prompt(
            "Enter API Base URL: ",
            default="https://api.openai.com/v1/",
        )

        model = typer.prompt(
            "Enter model name: ",
            default="text-davinci-003",
        )

        api_key = prompt(
            "Enter API key: ",
            is_password=True,
        )

        max_tokens = typer.prompt(
            "Enter max tokens",
            default=1024,
            type=int,
        )

        return ProviderConfig(
            provider=provider,
            api_base=api_base,
            model=model,
            api_key=api_key,
            max_tokens=max_tokens,
        )
    except KeyboardInterrupt as e:
        console.print(stylize("Operation cancelled by user.", Colors.MAGENTA))
        raise typer.Exit(1) from e


def entry(
    debug: Annotated[bool, typer.Option("--debug", "-d", help="Enable debug mode")] = False,
    local: Annotated[bool, typer.Option("--local", "-l", help="Use local config")] = False,
    config_path: Annotated[
        Optional[Path],
        typer.Option(
            "--config",
            "-c",
            help="Path to config file",
            exists=True,
            file_okay=True,
            dir_okay=False,
            writable=False,
            readable=True,
            resolve_path=True,
        ),
    ] = None,
) -> None:
    """
    Setup a new provider configuration.

    This command helps you configure a new AI provider with necessary settings
    like API base URL, model name, and authentication details.
    """
    if debug:
        set_debug()

    try:
        config_manager = (
            ConfigManager(config_path=config_path)
            if config_path
            else get_config_manager(local=local)
        )
        if debug:
            logger.debug(f"Using config file: {config_manager.current_config_path}")

        provider_config = create_provider_config()
        config_manager.set_new_provider(provider_config.provider, provider_config.to_dict())

        console.print(
            stylize(
                f"{GPTCOMET_PRE} Provider {provider_config.provider} configured successfully.",
                Colors.GREEN,
            )
        )

    except ConfigError as e:
        console.print(stylize(f"Configuration error: {e!s}", Colors.RED))
        raise typer.Exit(1) from None
    except Exception as e:
        console.print(stylize(f"Unexpected error: {e!s}", Colors.RED))
        if debug:
            logger.exception("Unexpected error occurred")
        raise typer.Exit(1) from None
