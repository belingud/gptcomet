import pytest

from gptcomet.utils import strtobool


def test_true_values():
    true_values = ["ok", "true", "yes", "1", "y", "on"]
    for value in true_values:
        assert strtobool(value) is True


def test_false_values():
    false_values = ["false", "no", "0", "n", "off"]
    for value in false_values:
        assert strtobool(value) is False


def test_invalid_values():
    invalid_values = ["hello", "123", "abc"]
    for value in invalid_values:
        with pytest.raises(ValueError):
            strtobool(value)


def test_case_insensitivity():
    assert strtobool("TRUE") is True
    assert strtobool("FALSE") is False
