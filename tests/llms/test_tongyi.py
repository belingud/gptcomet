"""Test TongyiLLM."""
import pytest

from gptcomet.llms.tongyi import TongyiLLM


@pytest.fixture
def tongyi_config(base_config):
    """Tongyi config fixture."""
    return base_config.copy()


def test_tongyi_init(tongyi_config):
    """Test TongyiLLM initialization."""
    llm = TongyiLLM(tongyi_config)
    assert llm.api_base == "https://dashscope.aliyuncs.com/compatible-mode/v1"
    assert llm.model == "qwen-turbo"


def test_tongyi_init_custom(tongyi_config):
    """Test TongyiLLM initialization with custom values."""
    config = tongyi_config.copy()
    config.update({
        "api_base": "https://custom.api.base",
        "model": "custom-model",
    })
    llm = TongyiLLM(config)
    assert llm.api_base == "https://custom.api.base"
    assert llm.model == "custom-model"


def test_tongyi_build_headers(tongyi_config):
    """Test build_headers method."""
    llm = TongyiLLM(tongyi_config)
    headers = llm.build_headers()
    assert headers["Authorization"] == "Bearer test-api-key"
    assert headers["Content-Type"] == "application/json"


def test_tongyi_build_payload_simple(tongyi_config, sample_message):
    """Test build_payload method with simple message."""
    llm = TongyiLLM(tongyi_config)
    payload = llm.format_messages(sample_message)

    assert payload["model"] == llm.model
    assert payload["max_tokens"] == tongyi_config["max_tokens"]
    assert payload["temperature"] == tongyi_config["temperature"]
    assert payload["top_p"] == tongyi_config["top_p"]
    assert len(payload["messages"]) == 1
    assert payload["messages"][0]["role"] == "user"
    assert payload["messages"][0]["content"] == sample_message


def test_tongyi_build_payload_with_history(tongyi_config, sample_message, sample_history):
    """Test build_payload method with chat history."""
    llm = TongyiLLM(tongyi_config)
    payload = llm.format_messages(sample_message, history=sample_history)

    assert len(payload["messages"]) == len(sample_history) + 1
    for i, msg in enumerate(sample_history):
        assert payload["messages"][i]["role"] == msg["role"]
        assert payload["messages"][i]["content"] == msg["content"]

    last_message = payload["messages"][-1]
    assert last_message["role"] == "user"
    assert last_message["content"] == sample_message


def test_tongyi_build_payload_with_penalties(tongyi_config, sample_message):
    """Test build_payload method with frequency and presence penalties."""
    config = tongyi_config.copy()
    config.update({
        "frequency_penalty": 0.5,
        "presence_penalty": 0.3,
    })

    llm = TongyiLLM(config)
    payload = llm.format_messages(sample_message)

    assert payload["frequency_penalty"] == 0.5
    assert payload["presence_penalty"] == 0.3


def test_tongyi_get_required_config():
    """Test get_required_config method."""
    config = TongyiLLM.get_required_config()
    assert isinstance(config, dict)
    assert "api_base" in config
    assert "model" in config
    assert "api_key" in config
    assert "max_tokens" in config

    # Check default values
    assert config["api_base"][0] == "https://dashscope.aliyuncs.com/compatible-mode/v1"
    assert config["model"][0] == "qwen-turbo"
