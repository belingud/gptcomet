"""Test OpenAI LLM."""
import pytest

from gptcomet.llms.openai import OpenaiLLM


@pytest.fixture
def openai_config(base_config):
    """OpenAI config fixture."""
    return base_config.copy()


def test_openai_init(openai_config):
    """Test OpenAI initialization."""
    llm = OpenaiLLM(openai_config)
    assert llm.api_base == "https://api.openai.com/v1"
    assert llm.model == "gpt-4o"
    assert llm.completion_path == "chat/completions"
    assert llm.answer_path == "choices.0.message.content"


def test_openai_init_custom(openai_config):
    """Test OpenAI initialization with custom values."""
    config = openai_config.copy()
    config.update({
        "api_base": "https://custom.api.base",
        "model": "custom-model",
        "completion_path": "custom/path",
        "answer_path": "custom.path",
    })
    llm = OpenaiLLM(config)
    assert llm.api_base == "https://custom.api.base"
    assert llm.model == "custom-model"
    assert llm.completion_path == "custom/path"
    assert llm.answer_path == "custom.path"


def test_openai_build_headers(openai_config):
    """Test build_headers method."""
    llm = OpenaiLLM(openai_config)
    headers = llm.build_headers()
    assert headers["Content-Type"] == "application/json"
    assert headers["Authorization"] == "Bearer test-api-key"


def test_openai_format_messages_simple(openai_config, sample_message):
    """Test format_messages method with simple message."""
    llm = OpenaiLLM(openai_config)
    payload = llm.format_messages(sample_message)

    assert payload["model"] == llm.model
    assert payload["max_tokens"] == openai_config["max_tokens"]
    assert payload["temperature"] == openai_config["temperature"]
    assert payload["top_p"] == openai_config["top_p"]
    assert len(payload["messages"]) == 1
    assert payload["messages"][0]["role"] == "user"
    assert payload["messages"][0]["content"] == sample_message


def test_openai_format_messages_with_history(openai_config, sample_message, sample_history):
    """Test format_messages method with chat history."""
    llm = OpenaiLLM(openai_config)
    payload = llm.format_messages(sample_message, history=sample_history)

    assert len(payload["messages"]) == len(sample_history) + 1
    for i, msg in enumerate(sample_history):
        assert payload["messages"][i]["role"] == msg["role"]
        assert payload["messages"][i]["content"] == msg["content"]

    last_message = payload["messages"][-1]
    assert last_message["role"] == "user"
    assert last_message["content"] == sample_message


def test_openai_format_messages_with_penalties(openai_config, sample_message):
    """Test format_messages method with frequency and presence penalties."""
    config = openai_config.copy()
    config.update({
        "frequency_penalty": 0.5,
        "presence_penalty": 0.3,
    })
    
    llm = OpenaiLLM(config)
    payload = llm.format_messages(sample_message)

    assert payload["frequency_penalty"] == 0.5
    assert payload["presence_penalty"] == 0.3


def test_openai_get_required_config():
    """Test get_required_config method."""
    config = OpenaiLLM.get_required_config()
    assert isinstance(config, dict)
    assert "api_base" in config
    assert "model" in config
    assert "api_key" in config
    assert "max_tokens" in config

    # Check default values
    assert config["api_base"][0] == "https://api.openai.com/v1"
    assert config["model"][0] == "gpt-4o"
