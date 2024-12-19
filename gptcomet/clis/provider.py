from pathlib import Path
from typing import Annotated, Optional

import typer
from prompt_toolkit import prompt
from prompt_toolkit.completion import WordCompleter

from gptcomet.config_manager import ConfigManager, ProviderConfig, get_config_manager
from gptcomet.const import GPTCOMET_PRE
from gptcomet.exceptions import ConfigError
from gptcomet.llms.providers import ProviderRegistry
from gptcomet.log import logger, set_debug
from gptcomet.styles import Colors, stylize
from gptcomet.utils import console, create_select_menu


def get_provider_name() -> str:
    """Get provider name from user with autocomplete support."""
    available_providers = list(ProviderRegistry._providers.keys())
    provider_completer = WordCompleter(available_providers, ignore_case=True)

    console.print(
        "You can either select one from the list or enter a custom provider name.",
    )
    try:
        provider = create_select_menu(available_providers)
        if provider != "INPUT_REQUIRED" and provider is not None:
            return provider
        provider = (
            prompt(
                "Enter provider name: ",
                completer=provider_completer,
                complete_while_typing=True,
            )
            .lower()
            .strip()
        )
        if not provider:
            console.print("Provider name cannot be empty.", style=Colors.RED)
            raise typer.Exit(1)
        else:
            return provider

    except KeyboardInterrupt as e:
        console.print(stylize("\nOperation cancelled by user.", Colors.MAGENTA))
        raise typer.Exit(1) from e


def create_provider_config() -> ProviderConfig:
    """Create provider config from user input."""
    # Get provider name with autocomplete
    provider = get_provider_name()

    # Get provider-specific configuration requirements
    provider_class = ProviderRegistry.get_provider(provider)
    config_requirements = provider_class.get_required_config()

    # Initialize config dict
    config_dict = {"provider": provider}

    # Get provider-specific configuration
    for key, (default_value, prompt_message) in config_requirements.items():
        if key == "api_key":
            value = prompt(
                f"{prompt_message}: ",
                is_password=True,
            )
            if not value:
                console.print("API key cannot be empty.", style=Colors.RED)
                raise typer.Exit(1)
        else:
            value = typer.prompt(
                prompt_message,
                default=default_value,
                type=str,
            )
            if not default_value and not value:
                console.print(f"{key} cannot be empty.", style=Colors.RED)
                raise typer.Exit(1)
        config_dict[key] = value

    # Convert to ProviderConfig
    return ProviderConfig(
        provider=provider,
        api_base=config_dict["api_base"],
        model=config_dict["model"],
        api_key=config_dict.get("api_key", ""),
        max_tokens=int(config_dict.get("max_tokens", 1024)),
        retries=int(config_dict.get("retries", 2)),
        **{
            k: v
            for k, v in config_dict.items()
            if k not in ["provider", "api_base", "model", "api_key", "max_tokens", "retries"]
        },
    )


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
    except KeyboardInterrupt:
        console.print(stylize("\nOperation cancelled by user.", Colors.MAGENTA))
        raise typer.Exit(1) from None
    except Exception as e:
        console.print(stylize(f"Unexpected error: {e!s}", Colors.RED))
        if debug:
            logger.exception("Unexpected error occurred")
        raise typer.Exit(1) from None
