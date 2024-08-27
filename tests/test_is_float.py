import pytest

from gptcomet.utils import is_float


def test_is_float_valid():
    assert is_float("3.14") is True
    assert is_float("123") is True
    assert is_float("-0.123") is True


def test_is_float_invalid():
    assert is_float("abc") is False
    assert is_float("123abc") is False
    assert is_float("abc123") is False


def test_is_float_empty():
    assert is_float("") is False


def test_is_float_non_string():
    with pytest.raises(TypeError):
        is_float(123)
    with pytest.raises(TypeError):
        is_float(None)
