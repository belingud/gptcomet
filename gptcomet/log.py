import logging

from rich.logging import RichHandler

logger = logging.getLogger("gptcomet")

formatter = logging.Formatter("%(message)s: ")
handler = RichHandler(level=logging.NOTSET, show_path=False)
handler.setFormatter(formatter)
logger.addHandler(handler)


def set_debug(debug=True):
    if not debug:
        return
    logger.setLevel(logging.DEBUG)
