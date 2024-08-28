import logging

logger = logging.getLogger("gptcomet")

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


def set_debug():
    set_level("DEBUG")
