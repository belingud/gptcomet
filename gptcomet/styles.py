import typing as t
from decimal import Decimal
from enum import StrEnum

from git import safe_decode
from rich.panel import Panel
from rich.text import Text

from gptcomet.utils import defenc


class Styles(StrEnum):
    BOLD = "bold"
    ITALIC = "italic"
    UNDERLINE = "underline"
    STRIKETHROUGH = "strikethrough"
    STRIKE = "strike"
    STRIKE_ITALIC = "strike italic"
    STRIKE_BOLD = "strike bold"
    STRIKE_BOLD_ITALIC = "strike bold italic"
    ITALIC_BOLD = "italic bold"
    ITALIC_BOLD_UNDERLINE = "italic bold underline"


class Colors(StrEnum):
    RED = "red"
    GREEN = "green"
    YELLOW = "yellow"
    BLUE = "blue"
    MAGENTA = "magenta"
    CYAN = "cyan"
    WHITE = "white"
    BLACK = "black"

    LIGHT_RED = "light_red"
    LIGHT_GREEN = "light_green"
    LIGHT_YELLOW = "light_yellow"
    LIGHT_BLUE = "light_blue"
    LIGHT_MAGENTA = "light_magenta"
    LIGHT_CYAN = "light_cyan"
    LIGHT_WHITE = "light_white"

    BRIGHT = "bright"

    DEFAULT = "default"


@t.overload
def stylize(text: str, *styles: str, panel: t.Literal[False] = ...) -> Text: ...


@t.overload
def stylize(text: str, *styles: str, panel: t.Literal[True] = ...) -> Panel: ...


@t.overload
def stylize(text: str, *styles: str) -> Text: ...


def stylize(text: str, *styles: str, panel: bool = False) -> t.Union[Text, Panel]:
    """
    Format the text with specified styles. If panel is True, the text will be formatted as a panel.

    Args:
        text (str): The text to format.
        *styles (str): The styles to apply to the text.
        panel (bool, optional): If True, the text will be formatted as a panel. Defaults to False.

    Returns:
        str: The formatted text.
    """
    styles = tuple(safe_decode(s) for s in styles)
    if isinstance(text, (int, float, Decimal)):
        text = str(text)
    elif isinstance(text, bytes):
        text = text.decode(defenc)
    t = Text(text, style=" ".join(styles))
    if panel is False:
        return t
    else:
        return Panel(t)
