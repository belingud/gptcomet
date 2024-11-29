import logging

import pytest

from gptcomet.log import LOG_LEVELS, logger, set_debug, set_level


def test_log_levels():
    # Test all valid log levels except NOTSET
    for level in [le for le in LOG_LEVELS if le != "NOTSET"]:
        set_level(level)
        assert logger.getEffectiveLevel() == getattr(logging, level)

    # Special test for NOTSET, which inherits parent logger's level
    set_level("NOTSET")
    assert logger.getEffectiveLevel() in [
        logging.WARNING,
        logging.NOTSET,
    ]  # Could be WARNING (root logger's default level) or NOTSET


def test_invalid_log_level():
    # Test invalid log levels
    with pytest.raises(ValueError) as exc:
        set_level("INVALID_LEVEL")
    assert "Invalid log level: INVALID_LEVEL" in str(exc.value)


def test_set_debug():
    # Test setting debug mode
    set_debug(True)
    assert logger.getEffectiveLevel() == logging.DEBUG

    # Test turning off debug mode (should not change current level)
    current_level = logger.getEffectiveLevel()
    set_debug(False)
    assert logger.getEffectiveLevel() == current_level


def test_logger_configuration():
    # Test basic logger configuration
    assert logger.name == "gptcomet"
    assert len(logger.handlers) == 1
    handler = logger.handlers[0]

    # Test handler configuration
    from rich.logging import RichHandler

    assert isinstance(handler, RichHandler)
    assert handler.level == logging.NOTSET

    # Test formatter
    formatter = handler.formatter
    assert formatter._fmt == "%(message)s: "
