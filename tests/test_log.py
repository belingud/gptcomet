import logging

import pytest

from gptcomet.log import LOG_LEVELS, logger, set_debug, set_level


def test_log_levels():
    # 测试所有有效的日志级别，除了 NOTSET
    for level in [l for l in LOG_LEVELS if l != "NOTSET"]:
        set_level(level)
        assert logger.getEffectiveLevel() == getattr(logging, level)

    # 特别测试 NOTSET，它会继承父 logger 的级别
    set_level("NOTSET")
    assert logger.getEffectiveLevel() in [logging.WARNING, logging.NOTSET]  # 可能是 WARNING（root logger 的默认级别）或 NOTSET


def test_invalid_log_level():
    # 测试无效的日志级别
    with pytest.raises(ValueError) as exc:
        set_level("INVALID_LEVEL")
    assert "Invalid log level: INVALID_LEVEL" in str(exc.value)


def test_set_debug():
    # 测试设置 debug 模式
    set_debug(True)
    assert logger.getEffectiveLevel() == logging.DEBUG

    # 测试关闭 debug 模式（不应改变当前级别）
    current_level = logger.getEffectiveLevel()
    set_debug(False)
    assert logger.getEffectiveLevel() == current_level


def test_logger_configuration():
    # 测试 logger 的基本配置
    assert logger.name == "gptcomet"
    assert len(logger.handlers) == 1
    handler = logger.handlers[0]
    
    # 测试 handler 的配置
    from rich.logging import RichHandler
    assert isinstance(handler, RichHandler)
    assert handler.level == logging.NOTSET
    
    # 测试 formatter
    formatter = handler.formatter
    assert formatter._fmt == "%(message)s: "
