import signal
import sys

import typer

from gptcomet import __version__
from gptcomet.clis import commit, config, provider
from gptcomet.utils import CONTEXT_SETTINGS


def ctrl_c_handler(sig, frame):
    print("User interrupted. Exiting...")
    sys.exit(1)


signal.signal(signal.SIGINT, ctrl_c_handler)
app = typer.Typer(
    name="gmsg",
    no_args_is_help=True,
    rich_markup_mode="rich",
    context_settings=CONTEXT_SETTINGS,
    pretty_exceptions_enable=False,
)

app.add_typer(config.app, name="config")
app.command("commit")(commit.entry)
app.command("newprovider")(provider.entry)


def version_callback(value: bool):
    if value:
        typer.echo(f"GPTComet Version: {__version__}")
        raise typer.Exit()


@app.callback()
def main(
    version: bool = typer.Option(None, "-v", "--version", callback=version_callback, is_eager=True),
):
    pass
