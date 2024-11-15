import logging
import time

logger = logging.getLogger(f"gptcomet.{__name__}")


def timeit_decorator(func):
    def wrapper(*args, **kwargs):
        start_time = time.perf_counter()
        result = func(*args, **kwargs)
        end_time = time.perf_counter()
        logger.debug(f"{func.__name__} executed in {end_time - start_time:.4f} seconds")
        return result

    return wrapper
