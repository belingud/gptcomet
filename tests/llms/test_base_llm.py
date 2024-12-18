from unittest.mock import Mock, patch

import httpx
import pytest
from rich.text import Text

from gptcomet.exceptions import ConfigError, ConfigErrorEnum, RequestError
from gptcomet.llms.base import BaseLLM


class _TestLLM(BaseLLM):
    """Test implementation of BaseLLM."""

    def format_messages(self, message: str, history=None):
        """Mock implementation of format_messages."""
        return {"messages": [{"content": message}]}


@pytest.fixture
def valid_config():
    """Fixture for valid LLM configuration."""
    return {
        "api_key": "test_key",
        "api_base": "https://api.test.com/v1/",
        "model": "test-model",
        "retries": 2,
        "timeout": 30,
        "proxy": "",
        "max_tokens": 100,
        "temperature": 0.7,
        "top_p": 1.0,
        "extra_headers": '{"X-Test": "test"}',
        "completion_path": "chat/completions",
        "answer_path": "choices.0.message.content",
    }


@pytest.fixture
def llm(valid_config):
    """Fixture for LLM instance."""
    return _TestLLM(valid_config)


def test_initialization(valid_config):
    """Test LLM initialization with valid config."""
    llm = _TestLLM(valid_config)
    assert llm.api_key == "test_key"
    assert llm.api_base == "https://api.test.com/v1"
    assert llm.model == "test-model"
    assert llm.retries == 2
    assert llm.timeout == 30
    assert llm.proxy == ""
    assert llm.max_tokens == 100
    assert llm.temperature == 0.7
    assert llm.top_p == 1.0
    assert llm.completion_path == "chat/completions"
    assert llm.answer_path == "choices.0.message.content"


def test_initialization_missing_api_key():
    """Test LLM initialization with missing API key."""
    config = {"api_base": "https://api.test.com"}
    with pytest.raises(ConfigError) as exc_info:
        _TestLLM(config)
    assert exc_info.value.error == ConfigErrorEnum.API_KEY_MISSING


def test_build_url(llm):
    """Test URL building."""
    url = llm.build_url()
    assert url == "https://api.test.com/v1/chat/completions"


def test_build_headers(llm):
    """Test header building."""
    headers = llm.build_headers()
    assert headers["Content-Type"] == "application/json"
    assert headers["Authorization"] == "Bearer test_key"
    assert headers["X-Test"] == "test"


def test_parse_response(llm):
    """Test response parsing."""
    # Test normal response
    response = {"choices": [{"message": {"content": "Hello world"}}]}
    assert llm.parse_response(response) == "Hello world"

    # Test code block response
    response = {"choices": [{"message": {"content": "```print('hello')```"}}]}
    assert llm.parse_response(response) == "print('hello')"

    # Test non-string response
    response = {"choices": [{"message": {"content": 42}}]}
    assert llm.parse_response(response) == 42


def test_get_usage(llm):
    """Test usage information formatting."""
    # Test with usage data
    data = {"usage": {"prompt_tokens": 10, "completion_tokens": 20, "total_tokens": 30}}
    usage = llm.get_usage(data)
    assert isinstance(usage, Text)
    assert "10" in str(usage)
    assert "20" in str(usage)
    assert "30" in str(usage)

    # Test without usage data
    assert llm.get_usage({}) is None


def test_make_request_success(llm):
    """Test successful API request."""
    mock_response = Mock()
    mock_response.json.return_value = {
        "choices": [{"message": {"content": "Hello world"}}],
        "usage": {"prompt_tokens": 10, "completion_tokens": 20, "total_tokens": 30},
    }

    with patch("httpx.Client.post", return_value=mock_response):
        result = llm.make_request("Hello")
        assert result == "Hello world"


def test_make_request_timeout(llm):
    """Test API request with timeout."""
    with patch("httpx.Client.post", side_effect=httpx.TimeoutException("Timeout")):
        with pytest.raises(RequestError) as exc_info:
            llm.make_request("Hello")
        assert "Request timed out" in str(exc_info.value)


def test_make_request_client_error(llm):
    """Test API request with client error (4xx)."""
    mock_response = Mock()
    mock_response.status_code = 400
    with patch(
        "httpx.Client.post",
        side_effect=httpx.HTTPStatusError("Client Error", request=Mock(), response=mock_response),
    ):
        with pytest.raises(RequestError) as exc_info:
            llm.make_request("Hello")
        assert "Client error" in str(exc_info.value)


def test_make_request_server_error(llm):
    """Test API request with server error (5xx)."""
    mock_response = Mock()
    mock_response.status_code = 500
    with patch(
        "httpx.Client.post",
        side_effect=httpx.HTTPStatusError("Server Error", request=Mock(), response=mock_response),
    ):
        with pytest.raises(RequestError) as exc_info:
            llm.make_request("Hello")
        assert "Server Error" in str(exc_info.value)


def test_client_lifecycle(llm):
    """Test HTTP client lifecycle management."""
    # Test lazy initialization
    assert llm._client is None
    client = llm.client
    assert client is not None
    assert isinstance(client, httpx.Client)

    # Test client reuse
    assert llm.client is client

    # Test client cleanup
    llm.__exit__(None, None, None)
    assert llm._client is None


def test_context_manager():
    """Test context manager interface."""
    config = {"api_key": "test_key"}
    with _TestLLM(config) as llm:
        assert isinstance(llm, _TestLLM)
        assert llm._client is None  # Client is lazily initialized
        client = llm.client
        assert isinstance(client, httpx.Client)
    assert llm._client is None  # Client is cleaned up
