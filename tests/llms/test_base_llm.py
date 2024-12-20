from unittest.mock import Mock, patch

import pytest
import requests
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
    assert llm.extra_headers == '{"X-Test": "test"}'
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
    assert llm.build_url() == "https://api.test.com/v1/chat/completions"


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
    data = {"usage": {"prompt_tokens": 10, "completion_tokens": 20, "total_tokens": 30}}
    usage = llm.get_usage(data)
    assert isinstance(usage, Text)
    assert "10" in str(usage)
    assert "20" in str(usage)
    assert "30" in str(usage)

    assert llm.get_usage({}) is None


def test_make_request_success(llm):
    """Test successful API request."""
    response = Mock()
    response.json.return_value = {"choices": [{"message": {"content": "test response"}}]}
    with patch("requests.Session.post", return_value=response):
        result = llm.make_request("test message")
        assert result == "test response"


def test_make_request_timeout(llm):
    """Test API request with timeout."""
    with patch("requests.Session.post", side_effect=requests.Timeout):
        with pytest.raises(RequestError) as exc_info:
            llm.make_request("test message")
        assert "Request timed out" in str(exc_info.value)


def test_make_request_client_error(llm):
    """Test API request with client error (4xx)."""
    response = Mock()
    response.status_code = 400
    response.raise_for_status.side_effect = requests.HTTPError(response=response)
    with patch("requests.Session.post", side_effect=response.raise_for_status):
        with pytest.raises(RequestError) as exc_info:
            llm.make_request("test message")
        assert "Request failed with status code 400" in str(exc_info.value)


def test_make_request_server_error(llm):
    """Test API request with server error (5xx)."""
    response = Mock()
    response.status_code = 500
    response.raise_for_status.side_effect = requests.HTTPError(response=response)
    with patch("requests.Session.post", side_effect=response.raise_for_status):
        with pytest.raises(RequestError) as exc_info:
            llm.make_request("test message")
        assert "Request failed with status code 500" in str(exc_info.value)


def test_client_lifecycle(llm):
    """Test HTTP client lifecycle management."""
    assert llm._session is None
    session = llm.session
    assert llm._session is not None
    assert session is llm._session

    # Test session reuse
    session2 = llm.session
    assert session2 is session

    # Test session cleanup
    llm.__del__()
    assert llm._session


def test_context_manager(valid_config):
    """Test context manager interface."""
    with _TestLLM(valid_config) as llm:
        assert llm._session is None
        session = llm.session
        assert llm._session is not None
    assert llm._session is None
