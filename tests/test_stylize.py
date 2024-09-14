from rich.panel import Panel
from rich.text import Text

from gptcomet.styles import stylize


def test_stylize_text():
    text = "Hello, world!"
    styles = ["bold", "italic"]
    result = stylize(text, *styles)
    assert isinstance(result, Text)
    assert result.style == "bold italic"
    assert result._text[0] == "Hello, world!"


def test_stylize_styles():
    text = "Hello, world!"
    styles = ["bold", "italic", "underline"]
    result = stylize(text, *styles)
    assert isinstance(result, Text)
    assert result.style == "bold italic underline"
    assert result._text[0] == "Hello, world!"


def test_stylize_int():
    text = 42
    styles = ["bold", "italic"]
    result = stylize(text, *styles)
    assert isinstance(result, Text)
    assert result.style == "bold italic"
    assert result._text[0] == "42"


def test_stylize_float():
    text = 3.14
    styles = ["bold", "italic"]
    result = stylize(text, *styles)
    assert isinstance(result, Text)
    assert result.style == "bold italic"
    assert result._text[0] == "3.14"


def test_stylize_bytes():
    text = b"Hello, world!"
    styles = ["bold", "italic"]
    result = stylize(text, *styles)
    assert isinstance(result, Text)
    assert result.style == "bold italic"
    assert result._text[0] == "Hello, world!"


def test_stylize_default_style():
    text = "Hello, world!"
    result = stylize(text)
    assert isinstance(result, Text)
    assert result.style == ""
    assert result._text[0] == "Hello, world!"
