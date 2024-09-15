import logging

from rich.logging import RichHandler

logger = logging.getLogger("gptcomet")

formatter = logging.Formatter(
    "%(message)s: "
)
handler = RichHandler(level=logging.NOTSET, show_path=False)
handler.setFormatter(formatter)
logger.addHandler(handler)

LOG_LEVELS = (
    "CRITICAL",
    "FATAL",
    "ERROR",
    "WARN",
    "WARNING",
    "INFO",
    "DEBUG",
    "NOTSET",
)


def set_level(level):
    if level not in LOG_LEVELS:
        msg = f"Invalid log level: {level}."
        raise ValueError(msg)
    logger.setLevel(level)


def set_debug(debug=True):
    if not debug:
        return
    set_level("DEBUG")
