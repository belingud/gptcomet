import typer

from gptcomet.clis import commit, config
from gptcomet.utils import CONTEXT_SETTINGS

app = typer.Typer(
    name="gmsg",
    no_args_is_help=True,
    rich_markup_mode="rich",
    context_settings=CONTEXT_SETTINGS,
    pretty_exceptions_enable=False,
)

app.add_typer(config.app, name="config")
app.command("commit")(commit.entry)

if __name__ == "__main__":
    app()
