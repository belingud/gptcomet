import typer

from gptcomet.clis.config import append, get, keys, path, remove, reset, set
from gptcomet.utils import CONTEXT_SETTINGS

from gptcomet.clis.config import list as _list  # isort:skip

app = typer.Typer(name="config", no_args_is_help=True, context_settings=CONTEXT_SETTINGS)

app.command(name="get", help="Get config value")(get.entry)
app.command(name="list", help="List config content")(_list.entry)
app.command(name="reset", help="Reset config to default")(reset.entry)
app.command(name="set", help="Set config value")(set.entry)
app.command(name="path", help="Get runtime config path")(path.entry)
app.command(name="keys", help="List supported config keys")(keys.entry)
app.command(
    name="append",
    help="Append a config value to sequence value, not modify if not exists",
    short_help="Append a config value",
)(append.entry)
app.command(
    name="remove",
    help="Remove a config value from sequence value, not modify if not exists",
    short_help="Remove a config value",
)(remove.entry)
