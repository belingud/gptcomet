import logging
import os
import sys

import click
from click import Context
from git import HookExecutionError, InvalidGitRepositoryError, NoSuchPathError
from litellm.exceptions import BadRequestError

import gptcomet
from gptcomet._validator import KEYS_VALIDATOR
from gptcomet.config_manager import ConfigManager
from gptcomet.const import GPTCOMET_PRE
from gptcomet.exceptions import (
    ConfigKeyError,
    ConfigKeyTypeError,
    GitNoStagedChanges,
    KeyNotFound,
    NotModified,
)
from gptcomet.hook import GPTCometHook
from gptcomet.message_generator import MessageGenerator
from gptcomet.utils import common_options

logger = logging.getLogger("gptcomet")
logger.setLevel(os.getenv("GPTCOMET_LOG_LEVEL", logging.WARNING))
handler = logging.StreamHandler(stream=sys.stdout)
handler.setFormatter(logging.Formatter(
    "%(asctime)s [%(levelname)s] %(filename)s::%(funcName)s: %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S",
))
logger.addHandler(handler)


cli: click.Group


@click.group(
    name="gptcomet",
    help="AI-Powered Git Commit Message Generator",
    context_settings={"help_option_names": ["-h", "--help"]},
)
@click.pass_context
@common_options
@click.version_option(
    gptcomet.__version__,
    "--version",
    "-v",
    prog_name="gptcomet",
    message="%(prog)s: %(version)s",
)
def cli(ctx: Context, debug, local):
    """GPTComet CLI"""
    if debug:
        logger.setLevel(logging.DEBUG)
    else:
        logger.setLevel(logging.INFO)
    logger.debug(f"Running with debug={debug}, local={local}")

    ctx.obj["config_manager"] = None


@cli.group(name="config")
@click.pass_context
@common_options
def config(ctx, debug, local):
    """Manage gptcomet configuration."""
    logger.debug(f"Config manage, local={ctx.obj['local']}")
    if not ctx.obj["config_manager"]:
        ctx.obj["config_manager"] = ConfigManager(
            config_path=ConfigManager.get_config_path(ctx.obj["local"])
        )


def _config_value_set_validate_callback(ctx, param, value):
    """
    Validate the value of the set command.

    Args:
        ctx (click.Context): The click context object.
        param (click.Parameter): The click parameter object.
        value (str): The value to be validated.

    Returns:
        str: The validated value.

    Raises:
        click.BadParameter: If the value is invalid.
    """
    key = ctx.params.get("key")
    msg = "Invalid value for key: "
    for k, v in KEYS_VALIDATOR.items():
        if k in key and not v["validator"](value):
            raise click.BadParameter(msg + "{}: {}".format(k, v["msg"]), ctx=ctx, param=param)
    return value


@config.command("set", help="Set a configuration value.")
@click.argument("key")
@click.argument("value", callback=_config_value_set_validate_callback)
@click.pass_context
@common_options
def config_set(ctx: Context, key, value, debug, local):
    """Set a configuration value."""
    config_manager: ConfigManager = ctx.obj["config_manager"]
    try:
        config_manager.set(key, value)
        click.echo(
            f"{GPTCOMET_PRE} Set {click.style(key, fg='green')} to {click.style(value, fg='green')}."
        )
    except ConfigKeyError as e:
        click.echo(f"{GPTCOMET_PRE} Error: {e!s}")


@config.command("append", help="Append a configuration value.")
@click.argument("key")
@click.argument("value")
@click.pass_context
@common_options
def config_append(ctx: Context, key, value, debug, local):
    """Append a configuration value."""
    try:
        ctx.obj["config_manager"].append(key, value)
        click.echo(
            f"{GPTCOMET_PRE} Appended {click.style(value, fg='green')} to {click.style(key, fg='green')}."
        )
    except NotModified:
        click.echo(f"{GPTCOMET_PRE} Config value already exists and not modified: {key!s}")
    except ConfigKeyTypeError as e:
        click.echo(f"{GPTCOMET_PRE} Error: {e!s}")


@config.command("remove", help="Remove a specific value from the list set by the corresponding key.")
@click.argument("key")
@click.argument("value")
@click.pass_context
@common_options
def config_remove(ctx: Context, key, value, debug, local):
    """Remove a configuration value."""
    config_manager: ConfigManager = ctx.obj["config_manager"]
    try:
        config_manager.remove(key, value)
        click.echo(
            f"{GPTCOMET_PRE} Removed {click.style(value, fg='green')} from {click.style(key, fg='green')}."
        )
    except ConfigKeyTypeError as e:
        click.echo(f"{GPTCOMET_PRE} Error: {e!s}")
    except ValueError:
        click.echo(f"{GPTCOMET_PRE} value not found: {value!s}")


@config.command("get")
@click.argument("key")
@click.pass_context
@common_options
def config_get(ctx, key, debug, local):
    """Get a configuration value."""
    config_manager: ConfigManager = ctx.obj["config_manager"]
    try:
        value = config_manager.get(key)
        click.echo(f"{key}: {value}")
    except ValueError as e:
        click.echo(f"Error: {e!s}")
    except ConfigKeyError as e:
        click.echo(str(e))


@config.command("list")
@click.pass_context
@common_options
def config_list(ctx: Context, debug, local):
    """List all configuration values."""
    click.echo(
        click.style("Current configuration:\n", fg="green") + ctx.obj["config_manager"].list()
    )


@config.command("reset")
@click.pass_context
@common_options
def reset(ctx: Context, debug, local):
    """Reset configuration to default values."""
    config_manager: ConfigManager = ctx.obj["config_manager"]
    config_manager.reset()
    click.echo("Configuration reset to default values")


@config.command(name="keys")
@click.pass_context
@common_options
def config_keys(ctx, debug, local):
    cfg_manager: ConfigManager = ctx.obj["config_manager"]
    keys = cfg_manager.list_keys()
    click.echo(click.style("Supported keys:\n", fg="green") + keys)


@config.command("path")
@click.pass_context
@common_options
def config_path(ctx: Context, debug, local):
    """Get the path to the configuration file."""
    config_manager: ConfigManager = ctx.obj["config_manager"]
    click.echo(
        click.style("Current configuration path:\n", fg="green")
        + config_manager.current_config_path.as_posix()
    )


@cli.group("hook", help="Manage GPTComet prepare-commit-msg hook.")
@common_options
@click.pass_context
def hook(ctx: click.Context, debug, local):
    pass


@hook.command("install", help="Install GPTComet prepare-commit-msg hook to current repository.")
@click.option(
    "--force/--no-force",
    "-f/-nf",
    default=False,
    help="Force installation or not.",
    show_default=True,
)
@common_options
@click.pass_context
def install_hook(ctx, debug, local, force=False):
    """Install GPTComet prepare-commit-msg hook."""
    try:
        comet_hook = GPTCometHook()
        if comet_hook.is_hook_installed() and not force:
            click.echo(
                "GPTComet prepare-commit-msg hook is already installed, use --force to force installation."
            )
            return
        comet_hook.install_hook()
        click.echo("GPTComet prepare-commit-msg hook has been installed successfully.")
    except InvalidGitRepositoryError as e:
        click.echo(f"Error: {e!s}")
    except Exception as e:
        click.echo(f"An error occurred while installing the hook: {e!s}")


@hook.command(
    "uninstall", help="Uninstall GPTComet prepare-commit-msg hook from current repository."
)
@common_options
@click.pass_context
def uninstall_hook(ctx: click.Context, debug, local, **kwargs):
    """Uninstall GPTComet prepare-commit-msg hook."""
    try:
        comet_hook = GPTCometHook()
        comet_hook.uninstall_hook()
    except (InvalidGitRepositoryError, NoSuchPathError) as e:
        click.echo(f"Error: {e!s}")
    except Exception as e:
        click.echo(f"An error occurred while uninstalling the hook: {e!s}")


@hook.command("status")
def hook_status():
    """Check if GPTComet prepare-commit-msg hook is installed."""
    try:
        comet_hook = GPTCometHook()
        if comet_hook.is_hook_installed():
            click.echo("GPTComet prepare-commit-msg hook is installed.")
        else:
            click.echo(
                f"GPTComet prepare-commit-msg hook is {click.style('not', fg='yellow')} installed."
            )
    except InvalidGitRepositoryError as e:
        click.echo(f"Error: {e!s}")
    except Exception as e:
        click.echo(f"An error occurred while checking the hook status: {e!s}")


@cli.group("gen", help="Generate a commit message based on `git diff --staged`.")
@click.pass_context
@common_options
def generate(ctx: click.Context, debug, local):
    if ctx.obj.get("config_manager") is None:
        ctx.obj["config_manager"] = ConfigManager(
            config_path=ConfigManager.get_config_path(ctx.obj["local"])
        )


@generate.command("commit")
# @click.option("--rich", is_flag=True, default=False, help="Generate rich commit message")
@common_options
@click.pass_context
def generate_commit(ctx: click.Context, debug, local, rich=False, **kwargs):
    """Generate a commit message based on git diff"""
    config_manager: ConfigManager = ctx.obj["config_manager"]
    message_generator = MessageGenerator(config_manager)
    retry = True
    commit_msg = None
    click.echo(
        click.style(
            "🤖 Hang tight! I'm having a chat with the AI to craft your commit message...",
            fg="cyan",
        )
    )
    while retry:
        try:
            commit_msg = message_generator.generate_commit_message(rich)
        except KeyNotFound as e:
            click.echo(f"Error: {e!s}, please check your configuration.")
            raise click.Abort() from None
        except (GitNoStagedChanges, BadRequestError) as e:
            click.echo(str(e))
            raise click.Abort() from None

        if commit_msg is None:
            click.echo(click.style("No commit message generated.", fg="magenta"))
            return

        click.echo("Generated commit message:")
        click.echo(click.style(commit_msg, fg="green"))

        if sys.stdin.isatty():
            # Interactive mode will ask for confirmation
            char = click.prompt(
                "Do you want to use this commit message? y: yes, n: no, r: retry.",
                default="y",
                type=click.Choice(["y", "n", "r"]),
            )
        else:
            click.echo(
                click.style(
                    "Non-interactive mode detected, using the generated commit message directly.",
                    fg="yellow",
                )
            )
            char = "y"

        if char == "n":
            click.echo(click.style("Commit message discarded.", fg="yellow"))
            return
        elif char == "y":
            retry = False
        logger.debug(f"Input: {char}")
    try:
        # message_generator.repo.index.commit(commit_msg)
        pass
    except (HookExecutionError, ValueError) as e:
        click.echo(f"Commit Error: {e!s}")
        raise click.Abort() from None
    click.echo(click.style("Commit message saved.", fg="green"))


@click.command()
@click.pass_context
def generate_prmsg(ctx: click.Context, debug, local):
    """Generate a pull request message based on changes compared to master"""
    config_manager: ConfigManager = ctx.obj["config_manager"]
    message_generator = MessageGenerator(config_manager)

    pr_msg = message_generator.generate_pr_message()

    click.echo("Generated pull request message:")
    click.echo(pr_msg)

    if click.confirm("Do you want to use this pull request message?"):
        # Here you could integrate with your Git hosting platform's API
        # to create a pull request with this message
        click.echo(
            "Pull request message saved. You can now use it to create a PR on your Git hosting platform."
        )
    else:
        click.echo("Pull request message discarded.")


if __name__ == "__main__":
    cli()
