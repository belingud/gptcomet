import typer

from gptcomet.clis.generate import commit
from gptcomet.utils import CONTEXT_SETTINGS

app = typer.Typer(name="gen", no_args_is_help=True, context_settings=CONTEXT_SETTINGS)

app.command(name="commit")(commit.entry)
