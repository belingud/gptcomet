from unittest.mock import Mock, patch

import httpx
import pytest

from gptcomet.exceptions import ConfigError, ConfigErrorEnum
from gptcomet.llm_client import LLMClient


@pytest.fixture
def mock_config_manager():
    """基础配置管理器mock"""
    config = Mock()
    config.get.side_effect = lambda key, default=None: {
        "provider": "openai",
        "openai.api_key": "test-key",
        "openai.model": "gpt-3.5-turbo",
        "openai.api_base": "https://api.openai.com/v1",
        "openai.retries": "3",
        "openai.completion_path": "/chat/completions",
        "openai.answer_path": "choices.0.message.content",
        "openai.proxy": "",
    }.get(key, default)
    config.is_api_key_set = True
    return config


@pytest.fixture
def llm_client(mock_config_manager):
    """基础LLMClient实例"""
    return LLMClient(mock_config_manager)


class TestLLMClientInitialization:
    def test_successful_initialization(self, mock_config_manager):
        """测试成功初始化"""
        client = LLMClient(mock_config_manager)
        assert client.provider == "openai"
        assert client.api_key == "test-key"
        assert client.model == "gpt-3.5-turbo"
        assert client.api_base == "https://api.openai.com/v1"
        assert client.retries == 3
        assert client.conversation_history == []

    def test_missing_api_key(self, mock_config_manager):
        """测试缺少API key的情况"""
        mock_config_manager.is_api_key_set = False
        with pytest.raises(ConfigError) as exc_info:
            LLMClient(mock_config_manager)
        assert exc_info.value.error == ConfigErrorEnum.API_KEY_MISSING

    def test_from_config_manager(self, mock_config_manager):
        """测试from_config_manager工厂方法"""
        client = LLMClient.from_config_manager(mock_config_manager)
        assert isinstance(client, LLMClient)
        assert client.api_key == "test-key"


class TestLLMClientGenerate:
    def test_generate_without_history(self, llm_client):
        """测试不使用历史记录的生成"""
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
        """测试使用历史记录的生成"""
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
        """测试响应中没有usage信息的情况"""
        mock_response = {"choices": [{"message": {"content": "Test response"}}]}

        with patch("gptcomet.llm_client.LLMClient.completion_with_retries") as mock_completion:
            mock_completion.return_value = mock_response
            response = llm_client.generate("Hello")
            assert response == "Test response"


class TestLLMClientHistory:
    def test_clear_history(self, llm_client):
        """测试清除历史记录"""
        llm_client.conversation_history = [
            {"role": "user", "content": "Hello"},
            {"role": "assistant", "content": "Hi"},
        ]
        llm_client.clear_history()
        assert llm_client.conversation_history == []


class TestLLMClientCompletionWithRetries:
    def test_successful_completion(self, llm_client):
        """测试成功的completion请求"""
        mock_response = {"choices": [{"message": {"content": "Test response"}}]}

        with patch("httpx.Client") as mock_client:
            mock_instance = Mock()
            mock_instance.post.return_value.json.return_value = mock_response
            mock_instance.post.return_value.raise_for_status.return_value = None
            mock_client.return_value = mock_instance

            response = llm_client.completion_with_retries(
                api_base="https://api.openai.com/v1",
                api_key="test-key",
                model="gpt-3.5-turbo",
                messages=[{"role": "user", "content": "Hello"}],
                max_tokens=100,
            )

            assert response == mock_response

    def test_completion_with_proxy(self, llm_client):
        """测试使用代理的completion请求"""
        llm_client.proxy = "http://proxy:8080"
        mock_response = {"choices": [{"message": {"content": "Test response"}}]}

        with patch("httpx.Client") as mock_client:
            mock_instance = Mock()
            mock_instance.post.return_value.json.return_value = mock_response
            mock_instance.post.return_value.raise_for_status.return_value = None
            mock_client.return_value = mock_instance

            llm_client.completion_with_retries(
                api_base="https://api.openai.com/v1",
                api_key="test-key",
                model="gpt-3.5-turbo",
                messages=[],
                max_tokens=100,
            )

            mock_client.assert_called_once()
            assert mock_client.call_args[1]["proxies"] == "http://proxy:8080"

    def test_completion_http_error(self, llm_client):
        """测试HTTP错误处理"""
        with patch("httpx.Client") as mock_client:
            mock_instance = Mock()
            mock_instance.post.side_effect = httpx.HTTPError("API Error")
            mock_client.return_value = mock_instance

            with pytest.raises(httpx.HTTPError):
                llm_client.completion_with_retries(
                    api_base="https://api.openai.com/v1",
                    api_key="test-key",
                    model="gpt-3.5-turbo",
                    messages=[],
                    max_tokens=100,
                )


class TestLLMClientGenChatParams:
    def test_gen_chat_params_basic(self, llm_client):
        """测试基本的chat参数生成"""
        messages = [{"role": "user", "content": "Hello"}]
        params = llm_client.gen_chat_params(messages)

        assert isinstance(params, dict)
        assert params["messages"] == messages
        assert params["model"] == "gpt-3.5-turbo"
        assert params["api_key"] == "test-key"

    def test_gen_chat_params_with_optional(self, mock_config_manager):
        """测试包含可选参数的chat参数生成"""
        mock_config_manager.get.side_effect = lambda key, default=None: {
            "provider": "openai",
            "openai.api_key": "test-key",
            "openai.model": "gpt-3.5-turbo",
            "openai.api_base": "https://api.openai.com/v1",
            "openai.max_tokens": "100",
            "openai.temperature": "0.7",
            "openai.top_p": "0.9",
            "openai.frequency_penalty": "0.5",
            "openai.extra_headers": '{"X-Custom": "value"}',
        }.get(key, default)

        client = LLMClient(mock_config_manager)
        params = client.gen_chat_params([])

        assert params["max_tokens"] == 100
        assert params["temperature"] == 0.7
        assert params["top_p"] == 0.9
        assert params["frequency_penalty"] == 0.5
        assert params["extra_headers"] == {"X-Custom": "value"}
