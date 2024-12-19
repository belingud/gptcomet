"""Test GeminiLLM."""
import pytest

from gptcomet.llms.gemini import GeminiLLM


@pytest.fixture
def gemini_config(base_config):
    """Gemini config fixture."""
    return base_config.copy()


def test_gemini_init(gemini_config):
    """Test GeminiLLM initialization."""
    llm = GeminiLLM(gemini_config)
    assert llm.api_base == "https://generativelanguage.googleapis.com/v1beta/models"
    assert llm.model == "gemini-pro"


def test_gemini_init_custom(gemini_config):
    """Test GeminiLLM initialization with custom values."""
    config = gemini_config.copy()
    config.update({
        "api_base": "https://custom.api.base",
        "model": "custom-model",
    })
    llm = GeminiLLM(config)
    assert llm.api_base == "https://custom.api.base"
    assert llm.model == "custom-model"


def test_gemini_build_headers(gemini_config):
    """Test build_headers method."""
    llm = GeminiLLM(gemini_config)
    headers = llm.build_headers()
    assert headers["Content-Type"] == "application/json"


def test_gemini_format_messages_simple(gemini_config, sample_message):
    """Test format_messages method with simple message."""
    llm = GeminiLLM(gemini_config)
    payload = llm.format_messages(sample_message)

    assert payload["contents"][0]["parts"][0]["text"] == sample_message
    assert "generationConfig" in payload
    assert payload["generationConfig"]["maxOutputTokens"] == gemini_config["max_tokens"]
    assert payload["generationConfig"]["temperature"] == gemini_config["temperature"]
    assert payload["generationConfig"]["topP"] == gemini_config["top_p"]


def test_gemini_format_messages_with_history(gemini_config, sample_message, sample_history):
    """Test format_messages method with chat history."""
    llm = GeminiLLM(gemini_config)
    payload = llm.format_messages(sample_message, history=sample_history)

    contents = payload["contents"]
    assert len(contents) == len(sample_history) + 1

    for i, msg in enumerate(sample_history):
        assert contents[i]["parts"][0]["text"] == msg["content"]

    last_message = contents[-1]
    assert last_message["role"] == "user"
    assert last_message["parts"][0]["text"] == sample_message


def test_gemini_get_required_config():
    """Test get_required_config method."""
    config = GeminiLLM.get_required_config()
    assert isinstance(config, dict)
    assert "api_base" in config
    assert "model" in config
    assert "api_key" in config
    assert "max_tokens" in config

    # Check default values
    assert config["api_base"][0] == "https://generativelanguage.googleapis.com/v1beta/models"
    assert config["model"][0] == "gemini-pro"
