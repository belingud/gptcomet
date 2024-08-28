import typer

from gptcomet.clis.config import get, path, reset, set
from gptcomet.clis.config import list as _list
from gptcomet.utils import CONTEXT_SETTINGS

app = typer.Typer(
    name="config", no_args_is_help=True, context_settings=CONTEXT_SETTINGS
)

app.command(name="get")(get.entry)
app.command(name="list")(_list.entry)
app.command(name="reset")(reset.entry)
app.command(name="set")(set.entry)
app.command(name="path")(path.entry)
