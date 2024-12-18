"""Test ClaudeLLM."""
import pytest

from gptcomet.llms.claude import ClaudeLLM


@pytest.fixture
def claude_config(base_config):
    """Claude config fixture."""
    config = base_config.copy()
    config.update({
        "anthropic-version": "2023-06-01",
        "top_k": 10,
    })
    return config


def test_claude_init(claude_config):
    """Test ClaudeLLM initialization."""
    llm = ClaudeLLM(claude_config)
    assert llm.api_base == "https://api.anthropic.com/v1"
    assert llm.model == "claude-3-5-sonnet"
    assert llm.completion_path == "messages"
    assert llm.answer_path == "content.0.text"
    assert llm.anthropic_version == "2023-06-01"
    assert llm.top_k == 10


def test_claude_init_defaults(base_config):
    """Test ClaudeLLM initialization with defaults."""
    llm = ClaudeLLM(base_config)
    assert llm.api_base == "https://api.anthropic.com/v1"
    assert llm.model == "claude-3-5-sonnet"
    assert llm.completion_path == "messages"
    assert llm.answer_path == "content.0.text"
    assert llm.anthropic_version == "2023-06-01"
    assert llm.top_k is None


def test_claude_build_headers(claude_config):
    """Test build_headers method."""
    llm = ClaudeLLM(claude_config)
    headers = llm.build_headers()
    assert headers["x-api-key"] == "test-api-key"
    assert headers["anthropic-version"] == "2023-06-01"
    assert headers["Content-Type"] == "application/json"


def test_claude_build_payload_simple(claude_config, sample_message):
    """Test build_payload method with simple message."""
    llm = ClaudeLLM(claude_config)
    payload = llm.format_messages(sample_message)

    assert payload["model"] == llm.model
    assert payload["max_tokens"] == claude_config["max_tokens"]
    assert payload["temperature"] == claude_config["temperature"]
    assert payload["top_p"] == claude_config["top_p"]
    assert payload["top_k"] == claude_config["top_k"]


def test_claude_get_required_config():
    """Test get_required_config method."""
    config = ClaudeLLM.get_required_config()
    assert isinstance(config, dict)
    assert "api_base" in config
    assert "model" in config
    assert "api_key" in config
    assert "max_tokens" in config

    # Check default values
    assert config["api_base"][0] == "https://api.anthropic.com/v1"
    assert config["model"][0] == "claude-3-5-sonnet"
