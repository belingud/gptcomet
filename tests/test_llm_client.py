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
                "extra_headers": '{"X-Custom": "value"}',
            },
            "prompt": {
                "system": "You are a helpful assistant.",
                "user": "Hello!",
            },
            "console.verbose": True,
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
        assert client.proxy == ""
        assert client.completion_path == "/chat/completions"
        assert client.content_path == "choices.0.message.content"

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
    @patch("gptcomet.llm_client.LLMClient.completion_with_retries")
    def test_generate_without_history(self, mock_completion, llm_client, mock_config_manager):
        """Test generation without history"""
        mock_response = {
            "choices": [{"message": {"content": "Test response"}}],
            "usage": {"prompt_tokens": 10, "completion_tokens": 20, "total_tokens": 30},
        }
        mock_completion.return_value = mock_response
        response = llm_client.generate("Hello", use_history=False)

        assert response == "Test response"
        assert len(llm_client.conversation_history) == 0
        mock_completion.assert_called_once()
        mock_completion.assert_called_with(
            model="gpt-3.5-turbo",
            api_key="test-key",
            api_base="https://api.openai.com/v1",
            messages=[{"role": "user", "content": "Hello"}],
            max_tokens=100,
            temperature=0.7,
            top_p=1.0,
            extra_headers={"X-Custom": "value"},
        )

    @patch("gptcomet.llm_client.LLMClient.completion_with_retries")
    def test_generate_with_history(self, mock_completion, llm_client):
        """Test generation with history"""
        mock_response = {
            "choices": [{"message": {"content": "Test response"}}],
            "usage": {"prompt_tokens": 10, "completion_tokens": 20, "total_tokens": 30},
        }
        mock_completion.return_value = mock_response
        response = llm_client.generate("Hello", use_history=True)

        assert response == "Test response"
        assert len(llm_client.conversation_history) == 2
        assert llm_client.conversation_history[-2]["role"] == "user"
        assert llm_client.conversation_history[-1]["role"] == "assistant"

    @patch("gptcomet.llm_client.LLMClient.completion_with_retries")
    def test_generate_without_usage_info(self, mock_completion, llm_client):
        """Test generation when response has no usage info"""
        mock_response = {"choices": [{"message": {"content": "Test response"}}]}
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
    def test_successful_completion(self, mock_config_manager):
        """Test successful completion request"""
        mock_response = {"choices": [{"message": {"content": "Test response"}}]}

        with patch("httpx.Client") as mock_client:
            llm_client = LLMClient(mock_config_manager)
            mock_instance = Mock(name="instance")
            mock_post_resp = Mock(name="post_response")
            mock_client.post.return_value = mock_post_resp
            mock_post_resp.json.return_value = mock_response
            # mock_post_resp.return_value = mock_post_resp
            # mock_instance.post.return_value.raise_for_status.return_value = None
            mock_client.return_value = mock_instance
            llm_client._http_client = mock_client

            response = llm_client.completion_with_retries(
                api_base="https://api.openai.com/v1",
                api_key="test-key",
                model="gpt-3.5-turbo",
                messages=[{"role": "user", "content": "Hello"}],
                max_tokens=100,
            )

            assert response == mock_response

    def test_completion_with_proxy(self, mock_config_manager):
        """Test completion request with proxy"""
        mock_config_manager.get.side_effect = lambda key, default=None: glom.glom(
            {
                "provider": "openai",
                "openai": {
                    "api_key": "test-key",
                    "model": "gpt-3.5-turbo",
                    "proxy": "http://proxy:8080",
                },
            },
            key,
            default=default,
        )
        mock_response = {"choices": [{"message": {"content": "Test response"}}]}

        with patch("httpx.Client") as mock_client:
            llm_client = LLMClient(mock_config_manager)
            mock_instance = Mock()
            mock_instance.post.return_value.json.return_value = mock_response
            mock_instance.post.return_value.raise_for_status.return_value = None
            mock_client.return_value = mock_instance
            llm_client._http_client = mock_client

            llm_client.completion_with_retries(
                api_base="https://api.openai.com/v1",
                api_key="test-key",
                model="gpt-3.5-turbo",
                messages=[],
                max_tokens=100,
            )

            mock_client.post.assert_called_once()
            assert mock_client.call_args[1]["proxy"] == "http://proxy:8080"

    def test_completion_http_error(self, llm_client):
        """Test HTTP error handling in completion_with_retries method"""
        with patch("httpx.Client") as mock_client:
            mock_response = Mock()
            mock_response = httpx.Response(status_code=500, json={"error": "Internal Server Error"})
            mock_response.request = Mock()
            mock_instance = Mock()
            mock_instance.post.return_value = mock_response
            mock_client.return_value = mock_instance

            llm_client._http_client = mock_client.return_value

            with pytest.raises(httpx.HTTPError):
                llm_client.completion_with_retries(
                    api_base="https://api.openai.com/v1",
                    api_key="test-key",
                    model="gpt-3.5-turbo",
                    messages=[],
                    max_tokens=100,
                )

            # 验证post方法被调用
            mock_instance.post.assert_called_once()

    def test_completion_with_retries_url_handling(self, mock_config_manager):
        """Test URL path handling"""
        mock_config_manager.get.side_effect = lambda key, default=None: glom.glom(
            {
                "provider": "openai",
                "openai": {
                    "api_key": "test-key",
                    "model": "gpt-3.5-turbo",
                    "api_base": "https://api.openai.com/",  # with trailing slash
                    "completion_path": "chat/completions",  # without trailing slash
                    "max_tokens": "100",
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
        with patch("httpx.Client") as mock_post:
            mock_post.post.return_value.json.return_value = {
                "choices": [{"message": {"content": "test response"}}]
            }
            mock_post.post.return_value.status_code = 200
            client._http_client = mock_post

            client.completion_with_retries(
                api_base="https://api.openai.com/",
                api_key="test-key",
                model="gpt-3.5-turbo",
                messages=[{"role": "user", "content": "test"}],
                max_tokens=100,
            )

            mock_post.post.assert_called_once()
            assert mock_post.post.call_args[0][0] == "https://api.openai.com/chat/completions"


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

            def close(self):
                pass

        import httpx

        original_client = httpx.Client
        httpx.Client = MockClient
        client = LLMClient(mock_config_manager)

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
                extra_headers={"X-Custom": "value"},
            )

            assert response == {"choices": [{"message": {"content": "test response"}}]}
        finally:
            httpx.Client = original_client
