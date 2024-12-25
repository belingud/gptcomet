import logging

from gptcomet.log import logger, set_debug


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
