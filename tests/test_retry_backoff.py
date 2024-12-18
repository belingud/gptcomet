from unittest.mock import Mock, patch

import httpx
import pytest

from gptcomet.llms.base import BaseLLM, RequestError, retry_with_backoff


class _TestLLM(BaseLLM):
    """Test LLM class for testing the retry_with_backoff decorator."""

    def __init__(self, retries=2):
        self.retries = retries
        super().__init__({"api_key": "test_key"})

    def format_messages(self, message: str, history=None):
        """Mock implementation of format_messages."""
        return {"messages": [{"content": message}]}

    @retry_with_backoff(base_delay=0.1, max_delay=0.3)
    def test_method(self, arg1, arg2=None):
        """Test method with retry decorator."""
        return f"{arg1}-{arg2}"

    mock_func = retry_with_backoff(base_delay=0.1, max_delay=0.3)(Mock())


def test_successful_call():
    """Test successful call without retries."""
    llm = _TestLLM()
    result = llm.test_method("test", arg2="value")
    assert result == "test-value"


def test_retry_on_timeout():
    """Test retry behavior on timeout exception."""
    calls = 0

    # Create a mock function that raises timeout then succeeds
    def mock_func(self, arg1, arg2=None):
        nonlocal calls
        calls += 1
        if calls == 1:
            raise httpx.TimeoutException("Timeout")
        return "success-None"

    original_mock_func = _TestLLM.mock_func
    _TestLLM.mock_func = retry_with_backoff(base_delay=0.1, max_delay=0.3)(mock_func)
    llm = _TestLLM(retries=1)

    result = llm.mock_func("success")
    assert result == "success-None"
    assert calls == 2

    # Restore the original mock function
    _TestLLM.mock_func = original_mock_func


def test_retry_on_http_error():
    """Test retry behavior on HTTP 5xx error."""
    calls = 0

    # Create a mock function that raises timeout then succeeds
    def mock_func(self, arg1, arg2=None):
        nonlocal calls
        calls += 1
        if calls == 1:
            raise httpx.HTTPStatusError(
                "Server Error", request=Mock(), response=Mock(status_code=503)
            )
        return "success-None"

    original_mock_func = _TestLLM.mock_func
    _TestLLM.mock_func = retry_with_backoff(base_delay=0.1, max_delay=0.3)(mock_func)
    llm = _TestLLM(retries=1)

    result = llm.mock_func("success")
    assert result == "success-None"
    assert calls == 2

    # Restore the original mock function
    _TestLLM.mock_func = original_mock_func


def test_no_retry_on_client_error():
    """Test no retry on HTTP 4xx client error."""
    calls = 0

    # Create a mock function that raises timeout then succeeds
    def mock_func(self, arg1, arg2=None):
        nonlocal calls
        calls += 1
        raise httpx.HTTPStatusError("Client Error", request=Mock(), response=Mock(status_code=400))

    original_mock_func = _TestLLM.mock_func
    _TestLLM.mock_func = retry_with_backoff(base_delay=0.1, max_delay=0.3)(mock_func)
    llm = _TestLLM(retries=2)

    with pytest.raises(RequestError) as exc_info:
        llm.mock_func("test")
    assert "Client error" in str(exc_info.value)
    assert calls == 1

    # Restore the original mock function
    _TestLLM.mock_func = original_mock_func


def test_max_retries_exceeded():
    """Test behavior when max retries are exceeded."""
    calls = 0

    # Create a mock function that raises timeout then succeeds
    def mock_func(self, arg1, arg2=None):
        nonlocal calls
        calls += 1
        raise Exception("Generic error")

    original_mock_func = _TestLLM.mock_func
    _TestLLM.mock_func = retry_with_backoff(base_delay=0.1, max_delay=0.3)(mock_func)
    llm = _TestLLM(retries=2)

    with pytest.raises(RequestError) as exc_info:
        llm.mock_func("test")
    assert "Unexpected error" in str(exc_info.value)
    assert calls == 3  # Initial try + 2 retries

    # Restore the original mock function
    _TestLLM.mock_func = original_mock_func


def test_backoff_timing():
    """Test exponential backoff timing."""
    calls = 0
    mock_sleep = Mock()

    # Create a mock function that raises timeout then succeeds
    def mock_func(self, arg1, arg2=None):
        nonlocal calls
        calls += 1
        if calls <= 2:
            raise Exception(f"Error {calls}")
        return "success-None"

    original_mock_func = _TestLLM.mock_func
    _TestLLM.mock_func = retry_with_backoff(base_delay=0.1, max_delay=0.3)(mock_func)
    llm = _TestLLM(retries=2)

    with patch("time.sleep", mock_sleep):
        result = llm.mock_func("test")
        assert result == "success-None"

        # Check sleep durations (0.1 * 2^attempt)
        mock_sleep.assert_any_call(0.1)  # First retry
        mock_sleep.assert_any_call(0.2)  # Second retry
        assert mock_sleep.call_count == 2

    # Restore the original mock function
    _TestLLM.mock_func = original_mock_func


def test_max_delay_cap():
    """Test that delay is capped at max_delay."""
    calls = 0
    mock_sleep = Mock()

    # Create a mock function that raises timeout then succeeds
    def mock_func(self, arg1, arg2=None):
        nonlocal calls
        calls += 1
        if calls <= 3:
            raise Exception(f"Error {calls}")
        return "success-None"

    original_mock_func = _TestLLM.mock_func
    _TestLLM.mock_func = retry_with_backoff(base_delay=0.1, max_delay=0.3)(mock_func)
    llm = _TestLLM(retries=3)

    with patch("time.sleep", mock_sleep):
        with pytest.raises(RequestError) as exc_info:
            llm.mock_func("test")
        assert "Unexpected error" in str(exc_info.value)

        # Verify that all sleep calls were <= max_delay
        for call in mock_sleep.call_args_list:
            assert call[0][0] <= 0.3

    # Restore the original mock function
    _TestLLM.mock_func = original_mock_func


def test_decorator_type_check():
    """Test that decorator can only be used with BaseLLM methods."""

    class NotLLM:
        @retry_with_backoff()
        def test_method(self):
            return "test"

    with pytest.raises(TypeError) as exc_info:
        NotLLM().test_method()
    assert "retry_with_backoff can only be used with BaseLLM methods" in str(exc_info.value)
