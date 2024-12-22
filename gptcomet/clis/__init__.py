import typer

from gptcomet import __version__
from gptcomet.clis import commit, config, provider
from gptcomet.utils import CONTEXT_SETTINGS

app = typer.Typer(
    name="gmsg",
    help="AI powered git commit message generator",
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
        typer.echo("GPTComet: AI powered git commit message generator")
        typer.echo(f"Version: {__version__}")
        raise typer.Exit()


@app.callback()
def main(
    version: bool = typer.Option(
        None,
        "-v",
        "--version",
        callback=version_callback,
        is_eager=True,
        help="Show the GPTComet version and exit.",
    ),
):
    pass
