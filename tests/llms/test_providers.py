import pytest

from gptcomet.exceptions import ConfigError, ConfigErrorEnum
from gptcomet.llms.claude import ClaudeLLM
from gptcomet.llms.cohere import CohereLLM
from gptcomet.llms.gemini import GeminiLLM
from gptcomet.llms.openai import OpenaiLLM


@pytest.fixture
def base_config():
    """Base configuration for all LLM providers."""
    return {
        "api_key": "test_key",
        "model": "test-model",
        "max_tokens": 100,
        "temperature": 0.7,
        "top_p": 0.9,
        "frequency_penalty": 0.2,
        "presence_penalty": 0.1,
        "completion_path": "/chat/completions",
    }


class TestGeminiLLM:
    """Test cases for GeminiLLM."""

    def test_initialization(self, base_config):
        """Test GeminiLLM initialization."""
        config = {**base_config, "completion_path": "generateContent"}
        llm = GeminiLLM(config)
        assert llm.api_base == "https://generativelanguage.googleapis.com/v1beta/models"
        assert llm.model == "test-model"
        assert llm.completion_path == "generateContent"
        assert llm.answer_path == "candidates.0.content.parts.0.text"

    def test_format_messages_no_history(self, base_config):
        """Test message formatting without history."""
        config = {**base_config, "completion_path": "generateContent"}
        llm = GeminiLLM(config)
        message = "Hello, world!"
        payload = llm.format_messages(message)

        assert payload["contents"][-1] == {"role": "user", "parts": [{"text": message}]}
        assert payload["generationConfig"]["maxOutputTokens"] == 100
        assert payload["generationConfig"]["temperature"] == 0.7
        assert payload["generationConfig"]["topP"] == 0.9
        assert payload["generationConfig"]["frequencyPenalty"] == 0.2
        assert payload["generationConfig"]["presencePenalty"] == 0.1

    def test_format_messages_with_history(self, base_config):
        """Test message formatting with history."""
        config = {**base_config, "completion_path": "generateContent"}
        llm = GeminiLLM(config)
        history = [{"role": "user", "content": "Hi"}, {"role": "assistant", "content": "Hello!"}]
        payload = llm.format_messages("How are you?", history)

        assert len(payload["contents"]) == 3
        assert payload["contents"][0] == {"role": "user", "parts": [{"text": "Hi"}]}
        assert payload["contents"][1] == {"role": "model", "parts": [{"text": "Hello!"}]}


class TestOpenaiLLM:
    """Test cases for OpenaiLLM."""

    def test_initialization(self, base_config):
        """Test OpenaiLLM initialization."""
        llm = OpenaiLLM(base_config)
        assert llm.api_base == "https://api.openai.com/v1"
        assert llm.model == "test-model"

    def test_format_messages_no_history(self, base_config):
        """Test message formatting without history."""
        llm = OpenaiLLM(base_config)
        message = "Hello, world!"
        payload = llm.format_messages(message)

        assert payload["messages"] == [{"role": "user", "content": message}]
        assert payload["model"] == "test-model"
        assert payload["max_tokens"] == 100
        assert payload["temperature"] == 0.7
        assert payload["top_p"] == 0.9
        assert payload["frequency_penalty"] == 0.2
        assert payload["presence_penalty"] == 0.1

    def test_format_messages_with_history(self, base_config):
        """Test message formatting with history."""
        llm = OpenaiLLM(base_config)
        history = [{"role": "user", "content": "Hi"}, {"role": "assistant", "content": "Hello!"}]
        payload = llm.format_messages("How are you?", history)

        assert len(payload["messages"]) == 3
        assert payload["messages"][0] == {"role": "user", "content": "Hi"}
        assert payload["messages"][1] == {"role": "assistant", "content": "Hello!"}


class TestCohereLLM:
    """Test cases for CohereLLM."""

    def test_initialization(self, base_config):
        """Test CohereLLM initialization."""
        llm = CohereLLM(base_config)
        assert llm.api_base == "https://api.cohere.ai/v1"
        assert llm.model == "test-model"

    def test_initialization_missing_api_key(self):
        """Test CohereLLM initialization with missing API key."""
        with pytest.raises(ConfigError) as exc_info:
            CohereLLM({"model": "test-model"})
        assert exc_info.value.error == ConfigErrorEnum.API_KEY_MISSING

    def test_format_messages_no_history(self, base_config):
        """Test message formatting without history."""
        llm = CohereLLM(base_config)
        message = "Hello, world!"
        payload = llm.format_messages(message)

        assert payload["message"] == message
        assert payload["model"] == "test-model"
        assert payload["max_tokens"] == 100
        assert payload["temperature"] == 0.7
        assert payload["top_p"] == 0.9
        assert payload["frequency_penalty"] == 0.2
        assert payload["presence_penalty"] == 0.1
        assert payload["chat_history"] == []

    def test_format_messages_with_history(self, base_config):
        """Test message formatting with history."""
        llm = CohereLLM(base_config)
        history = [{"role": "user", "content": "Hi"}, {"role": "assistant", "content": "Hello!"}]
        payload = llm.format_messages("How are you?", history)

        assert len(payload["chat_history"]) == 2
        assert payload["chat_history"][0] == {"role": "USER", "message": "Hi"}
        assert payload["chat_history"][1] == {"role": "CHATBOT", "message": "Hello!"}


class TestClaudeLLM:
    """Test cases for ClaudeLLM."""

    def test_initialization(self, base_config):
        """Test ClaudeLLM initialization."""
        config = {**base_config, "completion_path": "messages"}
        llm = ClaudeLLM(config)
        assert llm.api_base == "https://api.anthropic.com/v1"
        assert llm.model == "test-model"
        assert llm.completion_path == "messages"
        assert llm.answer_path == "content.0.text"

    def test_initialization_with_anthropic_version(self, base_config):
        """Test ClaudeLLM initialization with anthropic version."""
        config = {**base_config, "anthropic-version": "2023-06-01", "completion_path": "messages"}
        llm = ClaudeLLM(config)
        headers = llm.build_headers()
        assert headers["anthropic-version"] == "2023-06-01"

    def test_format_messages_no_history(self, base_config):
        """Test message formatting without history."""
        config = {**base_config, "completion_path": "messages"}
        llm = ClaudeLLM(config)
        message = "Hello, world!"
        payload = llm.format_messages(message)

        assert payload["prompt"] == "Hello, world!"
        assert payload["model"] == "test-model"
        assert payload["max_tokens"] == 100
        assert payload["temperature"] == 0.7
        assert payload["top_p"] == 0.9

    def test_format_messages_with_history(self, base_config):
        """Test message formatting with history."""
        config = {**base_config, "completion_path": "messages"}
        llm = ClaudeLLM(config)
        history = [{"role": "user", "content": "Hi"}, {"role": "assistant", "content": "Hello!"}]
        payload = llm.format_messages("How are you?", history)

        expected_prompt = "How are you?"
        assert payload["prompt"] == expected_prompt
