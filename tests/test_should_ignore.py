from gptcomet.utils import should_ignore


def test_should_ignore_empty_ignore_patterns():
    filepath = "/path/to/file.txt"
    ignore_patterns = []
    assert should_ignore(filepath, ignore_patterns) is False


def test_should_ignore_not_match():
    filepath = "/path/to/file.txt"
    ignore_patterns = ["*.py"]
    assert should_ignore(filepath, ignore_patterns) is False


def test_should_ignore_match():
    filepath = "/path/to/file.py"
    ignore_patterns = ["*.py"]
    assert should_ignore(filepath, ignore_patterns) is True


def test_should_ignore_multiple_match():
    filepath = "/path/to/file.py"
    ignore_patterns = ["*.py", "*.txt"]
    assert should_ignore(filepath, ignore_patterns) is True
