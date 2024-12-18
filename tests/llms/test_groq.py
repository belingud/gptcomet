"""Test Groq LLM."""
import pytest

from gptcomet.llms.groq import GroqLLM


@pytest.fixture
def groq_config(base_config):
    """Groq config fixture."""
    return base_config.copy()


def test_groq_init(groq_config):
    """Test Groq initialization."""
    llm = GroqLLM(groq_config)
    assert llm.api_base == "https://api.groq.com/openai/v1"
    assert llm.model == "llama3-8b-8192"


def test_groq_init_custom(groq_config):
    """Test Groq initialization with custom values."""
    config = groq_config.copy()
    config.update({
        "api_base": "https://custom.groq.api",
        "model": "custom-model",
    })
    llm = GroqLLM(config)
    assert llm.api_base == "https://custom.groq.api"
    assert llm.model == "custom-model"


def test_groq_build_headers(groq_config):
    """Test build_headers method."""
    llm = GroqLLM(groq_config)
    headers = llm.build_headers()
    assert headers["Content-Type"] == "application/json"
    assert headers["Authorization"] == "Bearer test-api-key"


def test_groq_format_messages_simple(groq_config, sample_message):
    """Test format_messages method with simple message."""
    llm = GroqLLM(groq_config)
    payload = llm.format_messages(sample_message)

    assert payload["model"] == llm.model
    assert payload["max_tokens"] == groq_config["max_tokens"]
    assert payload["temperature"] == groq_config["temperature"]
    assert payload["top_p"] == groq_config["top_p"]
    assert len(payload["messages"]) == 1
    assert payload["messages"][0]["role"] == "user"
    assert payload["messages"][0]["content"] == sample_message


def test_groq_format_messages_with_history(groq_config, sample_message, sample_history):
    """Test format_messages method with chat history."""
    llm = GroqLLM(groq_config)
    payload = llm.format_messages(sample_message, history=sample_history)

    assert len(payload["messages"]) == len(sample_history) + 1
    for i, msg in enumerate(sample_history):
        assert payload["messages"][i]["role"] == msg["role"]
        assert payload["messages"][i]["content"] == msg["content"]

    last_message = payload["messages"][-1]
    assert last_message["role"] == "user"
    assert last_message["content"] == sample_message


def test_groq_get_required_config():
    """Test get_required_config method."""
    config = GroqLLM.get_required_config()
    assert isinstance(config, dict)
    assert "api_base" in config
    assert "model" in config
    assert "api_key" in config
    assert "max_tokens" in config

    # Check default values
    assert config["api_base"][0] == "https://api.groq.com/openai/v1"
    assert config["model"][0] == "llama3-8b-8192"
