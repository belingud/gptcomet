from unittest.mock import Mock, patch

import glom
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
        assert client.provider_name == "openai"
        assert client.model == "gpt-3.5-turbo"
        assert client.retries == 3
        assert client.conversation_history == []

    def test_missing_api_key(self, mock_config_manager):
        """Test initialization without API key"""
        mock_config_manager.get.side_effect = lambda key, default=None: glom.glom(
            {
                "provider": "openai",
                "openai": {
                    "model": "gpt-3.5-turbo",
                },
            },
            key,
            default=default,
        )
        with pytest.raises(ConfigError) as exc_info:
            LLMClient(mock_config_manager)
        assert exc_info.value.error == ConfigErrorEnum.API_KEY_MISSING
        assert exc_info.value.provider == "openai"

    def test_from_config_manager(self, mock_config_manager):
        """Test from_config_manager factory method"""
        client = LLMClient.from_config_manager(mock_config_manager)
        assert isinstance(client, LLMClient)
        assert client.provider_name == "openai"
        assert client.model == "gpt-3.5-turbo"

    def test_provider_config_missing(self, mock_config_manager):
        """Test missing provider configuration"""
        mock_config_manager.get.side_effect = lambda key, default=None: glom.glom(
            {
                "provider": None,
            },
            key,
            default=default,
        )

        with pytest.raises(ConfigError) as exc_info:
            LLMClient(mock_config_manager)
        assert exc_info.value.error == ConfigErrorEnum.PROVIDER_KEY_MISSING


class TestLLMClientGenerate:
    @patch("gptcomet.llm_client.LLMClient.completion_with_retries")
    def test_generate_without_history(self, mock_completion, llm_client, mock_config_manager):
        """Test generation without history"""
        mock_completion.return_value = "Test response"
        response = llm_client.generate("Hello", use_history=False)

        assert response == "Test response"
        assert len(llm_client.conversation_history) == 0
        mock_completion.assert_called_once()
        mock_completion.assert_called_with("Hello")

    @patch("gptcomet.llm_client.LLMClient.completion_with_retries")
    def test_generate_with_history(self, mock_completion, llm_client):
        """Test generation with history"""
        mock_completion.return_value = "Test response"
        llm_client.conversation_history = [
            {"role": "user", "content": "Previous message"},
            {"role": "assistant", "content": "Previous response"},
        ]
        response = llm_client.generate("Hello", use_history=True)

        assert response == "Test response"
        mock_completion.assert_called_once()
        mock_completion.assert_called_with("Hello", history=llm_client.conversation_history)


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
        expected_response = "Test response"

        with patch("gptcomet.llms.base.BaseLLM.make_request") as mock_make_request:
            llm_client = LLMClient(mock_config_manager)
            mock_make_request.return_value = expected_response

            response = llm_client.completion_with_retries("Hello")
            assert response == expected_response
            mock_make_request.assert_called_once_with("Hello", None)

    def test_completion_retry_on_failure(self, mock_config_manager):
        """Test completion request with retries on failure"""
        expected_response = "Test response"

        with patch("gptcomet.llms.base.BaseLLM.make_request") as mock_make_request:
            llm_client = LLMClient(mock_config_manager)
            mock_make_request.side_effect = [Exception("API Error"), expected_response]

            response = llm_client.completion_with_retries("Hello")
            assert response == expected_response
            assert mock_make_request.call_count == 2
            mock_make_request.assert_called_with("Hello", None)

    def test_completion_failure_after_retries(self, mock_config_manager):
        """Test completion request fails after all retries"""
        with patch("gptcomet.llms.base.BaseLLM.make_request") as mock_make_request:
            llm_client = LLMClient(mock_config_manager)
            mock_make_request.side_effect = Exception("API Error")

            with pytest.raises(Exception) as exc_info:
                llm_client.completion_with_retries("Hello")
            assert str(exc_info.value) == "API Error"
            assert mock_make_request.call_count == llm_client.retries + 1
