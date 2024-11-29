from unittest.mock import Mock, patch

import glom
import httpx
import pytest

from gptcomet.exceptions import ConfigError, ConfigErrorEnum
from gptcomet.llm_client import LLMClient


@pytest.fixture
def mock_config_manager():
    """Basic config manager mock"""
    config = Mock()
    config.get.side_effect = lambda key, default=None: glom.glom(
        {
            "provider": "openai",
            "openai": {
                "api_key": "test-key",
                "model": "gpt-3.5-turbo",
                "api_base": "https://api.openai.com/v1",
                "retries": "3",
                "completion_path": "/chat/completions",
                "answer_path": "choices.0.message.content",
                "proxy": "",
                "max_tokens": "100",
                "temperature": "0.7",
                "top_p": "1",
                "frequency_penalty": "0",
                "presence_penalty": "0",
            },
            "prompt": {
                "system": "You are a helpful assistant.",
                "user": "Hello!",
            },
        },
        key,
        default=default,
    )
    config.is_api_key_set = True
    return config


@pytest.fixture
def llm_client(mock_config_manager):
    """Basic LLMClient instance"""
    return LLMClient(mock_config_manager)


class TestLLMClientInitialization:
    def test_successful_initialization(self, mock_config_manager):
        """Test successful initialization"""
        client = LLMClient(mock_config_manager)
        assert client.provider == "openai"
        assert client.api_key == "test-key"
        assert client.model == "gpt-3.5-turbo"
        assert client.api_base == "https://api.openai.com/v1"
        assert client.retries == 3
        assert client.conversation_history == []

    def test_missing_api_key(self, mock_config_manager):
        """Test initialization without API key"""
        mock_config_manager.is_api_key_set = False
        with pytest.raises(ConfigError) as exc_info:
            LLMClient(mock_config_manager)
        assert exc_info.value.error == ConfigErrorEnum.API_KEY_MISSING

    def test_from_config_manager(self, mock_config_manager):
        """Test from_config_manager factory method"""
        client = LLMClient.from_config_manager(mock_config_manager)
        assert isinstance(client, LLMClient)
        assert client.api_key == "test-key"

    def test_provider_config_missing(self, mock_config_manager):
        """Test missing provider configuration"""
        mock_config_manager.get.side_effect = lambda key, default=None: glom.glom(
            {
                "provider": "openai",
            },
            key,
            default=default,
        )

        with pytest.raises(ConfigError) as exc_info:
            LLMClient(mock_config_manager)
        assert exc_info.value.error == ConfigErrorEnum.PROVIDER_CONFIG_MISSING


class TestLLMClientGenerate:
    def test_generate_without_history(self, llm_client):
        """Test generation without history"""
        mock_response = {
            "choices": [{"message": {"content": "Test response"}}],
            "usage": {"prompt_tokens": 10, "completion_tokens": 20, "total_tokens": 30},
        }

        with patch("gptcomet.llm_client.LLMClient.completion_with_retries") as mock_completion:
            mock_completion.return_value = mock_response
            response = llm_client.generate("Hello", use_history=False)

            assert response == "Test response"
            assert len(llm_client.conversation_history) == 0
            mock_completion.assert_called_once()

    def test_generate_with_history(self, llm_client):
        """Test generation with history"""
        mock_response = {
            "choices": [{"message": {"content": "Test response"}}],
            "usage": {"prompt_tokens": 10, "completion_tokens": 20, "total_tokens": 30},
        }

        with patch("gptcomet.llm_client.LLMClient.completion_with_retries") as mock_completion:
            mock_completion.return_value = mock_response
            response = llm_client.generate("Hello", use_history=True)

            assert response == "Test response"
            assert len(llm_client.conversation_history) == 2
            assert llm_client.conversation_history[-2]["role"] == "user"
            assert llm_client.conversation_history[-1]["role"] == "assistant"

    def test_generate_without_usage_info(self, llm_client):
        """Test generation when response has no usage info"""
        mock_response = {"choices": [{"message": {"content": "Test response"}}]}

        with patch("gptcomet.llm_client.LLMClient.completion_with_retries") as mock_completion:
            mock_completion.return_value = mock_response
            response = llm_client.generate("Hello")
            assert response == "Test response"


class TestLLMClientHistory:
    def test_clear_history(self, llm_client):
        """Test clearing history"""
        llm_client.conversation_history = [
            {"role": "user", "content": "Hello"},
            {"role": "assistant", "content": "Hi"},
        ]
        llm_client.clear_history()
        assert llm_client.conversation_history == []


class TestLLMClientCompletionWithRetries:
    def test_successful_completion(self, llm_client):
        """Test successful completion request"""
        mock_response = {"choices": [{"message": {"content": "Test response"}}]}

        with patch.object(llm_client._http_client, 'post') as mock_post:
            mock_post.return_value.json.return_value = mock_response
            mock_post.return_value.raise_for_status.return_value = None

            response = llm_client.completion_with_retries(
                api_base="https://api.openai.com/v1",
                api_key="test-key",
                model="gpt-3.5-turbo",
                messages=[{"role": "user", "content": "Hello"}],
                max_tokens=100,
            )

            assert response == mock_response
            mock_post.assert_called_once()

    def test_completion_with_optional_params(self, llm_client):
        """Test completion request with optional parameters"""
        mock_response = {"choices": [{"message": {"content": "Test response"}}]}

        with patch.object(llm_client._http_client, 'post') as mock_post:
            mock_post.return_value.json.return_value = mock_response
            mock_post.return_value.raise_for_status.return_value = None

            response = llm_client.completion_with_retries(
                api_base="https://api.openai.com/v1",
                api_key="test-key",
                model="gpt-3.5-turbo",
                messages=[{"role": "user", "content": "Hello"}],
                max_tokens=100,
                temperature=0.7,
                top_p=0.9,
                frequency_penalty=0.5,
            )

            assert response == mock_response
            mock_post.assert_called_once()
            
            # Verify optional parameters in payload
            payload = mock_post.call_args[1]["json"]
            assert payload["temperature"] == 0.7
            assert payload["top_p"] == 0.9
            assert payload["frequency_penalty"] == 0.5

    def test_completion_with_extra_headers(self, llm_client):
        """Test completion request with extra headers"""
        mock_response = {"choices": [{"message": {"content": "Test response"}}]}
        extra_headers = {"X-Custom-Header": "test-value"}

        with patch.object(llm_client._http_client, 'post') as mock_post:
            mock_post.return_value.json.return_value = mock_response
            mock_post.return_value.raise_for_status.return_value = None

            response = llm_client.completion_with_retries(
                api_base="https://api.openai.com/v1",
                api_key="test-key",
                model="gpt-3.5-turbo",
                messages=[{"role": "user", "content": "Hello"}],
                max_tokens=100,
                extra_headers=extra_headers,
            )

            assert response == mock_response
            mock_post.assert_called_once()
            
            # Verify headers
            headers = mock_post.call_args[1]["headers"]
            assert headers["X-Custom-Header"] == "test-value"
            assert headers["Authorization"] == "Bearer test-key"
            assert headers["Content-Type"] == "application/json"

    def test_completion_http_error(self, llm_client):
        """Test HTTP error handling"""
        with patch.object(llm_client._http_client, 'post') as mock_post:
            mock_post.side_effect = httpx.HTTPError("API Error")

            with pytest.raises(httpx.HTTPError):
                llm_client.completion_with_retries(
                    api_base="https://api.openai.com/v1",
                    api_key="test-key",
                    model="gpt-3.5-turbo",
                    messages=[{"role": "user", "content": "Hello"}],
                    max_tokens=100,
                )

    def test_completion_timeout_error(self, llm_client):
        """Test timeout error handling"""
        with patch.object(llm_client._http_client, 'post') as mock_post:
            mock_post.side_effect = httpx.TimeoutException("Request timed out")

            with pytest.raises(httpx.TimeoutException):
                llm_client.completion_with_retries(
                    api_base="https://api.openai.com/v1",
                    api_key="test-key",
                    model="gpt-3.5-turbo",
                    messages=[{"role": "user", "content": "Hello"}],
                    max_tokens=100,
                )

    def test_completion_url_handling(self, llm_client):
        """Test URL path handling"""
        mock_response = {"choices": [{"message": {"content": "Test response"}}]}

        with patch.object(llm_client._http_client, 'post') as mock_post:
            mock_post.return_value.json.return_value = mock_response
            mock_post.return_value.raise_for_status.return_value = None

            # Test with trailing slash in api_base
            llm_client.completion_with_retries(
                api_base="https://api.openai.com/",
                api_key="test-key",
                model="gpt-3.5-turbo",
                messages=[{"role": "user", "content": "Hello"}],
                max_tokens=100,
            )

            mock_post.assert_called_once()
            assert mock_post.call_args[0][0].endswith("/chat/completions")
            assert not mock_post.call_args[0][0].endswith("//chat/completions")


class TestLLMClientGenChatParams:
    def test_gen_chat_params_basic(self, llm_client):
        """Test basic chat parameter generation"""
        messages = [{"role": "user", "content": "Hello"}]
        params = llm_client.gen_chat_params(messages)

        assert isinstance(params, dict)
        assert params["messages"] == messages
        assert params["model"] == "gpt-3.5-turbo"
        assert params["api_key"] == "test-key"

    def test_gen_chat_params_with_optional(self, mock_config_manager):
        """Test chat parameter generation with optional parameters"""
        mock_config_manager.get.side_effect = lambda key, default=None: glom.glom(
            {
                "provider": "openai",
                "openai": {
                    "api_key": "test-key",
                    "model": "gpt-3.5-turbo",
                    "api_base": "https://api.openai.com/v1",
                    "max_tokens": "100",
                    "temperature": "0.7",
                    "top_p": "0.9",
                    "frequency_penalty": "0.5",
                    "extra_headers": '{"X-Custom": "value"}',
                },
                "prompt": {
                    "system": "You are a helpful assistant.",
                    "user": "Hello!",
                },
            },
            key,
            default=default,
        )

        client = LLMClient(mock_config_manager)
        params = client.gen_chat_params([])

        assert params["max_tokens"] == 100
        assert params["temperature"] == 0.7
        assert params["top_p"] == 0.9
        assert params["frequency_penalty"] == 0.5
        assert params["extra_headers"] == {"X-Custom": "value"}

    def test_gen_chat_params_with_invalid_extra_headers(self, mock_config_manager):
        """Test chat parameter generation with invalid extra headers"""
        mock_config_manager.get.side_effect = lambda key, default=None: glom.glom(
            {
                "provider": "openai",
                "openai": {
                    "api_key": "test-key",
                    "model": "gpt-3.5-turbo",
                    "api_base": "https://api.openai.com/v1",
                    "max_tokens": "100",
                    "temperature": "0.7",
                    "top_p": "0.9",
                    "frequency_penalty": "0.5",
                    "extra_headers": "{invalid json}",
                },
                "prompt": {
                    "system": "You are a helpful assistant.",
                    "user": "Hello!",
                },
            },
            key,
            default=default,
        )

        client = LLMClient(mock_config_manager)
        params = client.gen_chat_params([])

        assert "extra_headers" not in params

    def test_gen_chat_params_with_all_optional_params(self, mock_config_manager):
        """Test chat parameter generation with all optional parameters"""
        mock_config_manager.get.side_effect = lambda key, default=None: glom.glom(
            {
                "provider": "openai",
                "openai": {
                    "api_key": "test-key",
                    "model": "gpt-3.5-turbo",
                    "temperature": "0.7",
                    "top_p": "0.9",
                    "frequency_penalty": "0.5",
                },
            },
            key,
            default=default,
        )
        mock_config_manager.is_api_key_set = True

        client = LLMClient(mock_config_manager)
        params = client.gen_chat_params([{"role": "user", "content": "test"}])

        assert params["temperature"] == 0.7
        assert params["top_p"] == 0.9
        assert params["frequency_penalty"] == 0.5

    def test_completion_with_retries_with_all_optional_params(self, mock_config_manager):
        """Test completion_with_retries with all optional parameters"""
        mock_config_manager.get.side_effect = lambda key, default=None: glom.glom(
            {"provider": "openai", "openai": {"api_key": "test-key", "model": "gpt-3.5-turbo"}},
            key,
            default=default,
        )
        mock_config_manager.is_api_key_set = True

        client = LLMClient(mock_config_manager)

        # Mock httpx.Client
        class MockResponse:
            def __init__(self):
                self.status_code = 200
                self.text = '{"choices": [{"message": {"content": "test response"}}]}'

            def raise_for_status(self):
                pass

            def json(self):
                return {"choices": [{"message": {"content": "test response"}}]}

        class MockClient:
            def __init__(self, *args, **kwargs):
                pass

            def post(self, *args, **kwargs):
                return MockResponse()

            def __enter__(self):
                return self

            def __exit__(self, *args):
                pass

        import httpx

        original_client = httpx.Client
        httpx.Client = MockClient

        try:
            response = client.completion_with_retries(
                api_base="https://api.openai.com/v1",
                api_key="test-key",
                model="gpt-3.5-turbo",
                messages=[{"role": "user", "content": "test"}],
                max_tokens=100,
                temperature=0.7,
                top_p=0.9,
                frequency_penalty=0.5,
            )

            assert response == {"choices": [{"message": {"content": "test response"}}]}
        finally:
            httpx.Client = original_client
