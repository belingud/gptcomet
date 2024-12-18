"""Test ChatGLM LLM."""
import pytest

from gptcomet.llms.chatglm import ChatGLMLLM


@pytest.fixture
def chatglm_config(base_config):
    """ChatGLM config fixture."""
    return base_config.copy()


def test_chatglm_init(chatglm_config):
    """Test ChatGLM initialization."""
    llm = ChatGLMLLM(chatglm_config)
    assert llm.api_base == "https://open.bigmodel.cn/api/paas/v4"
    assert llm.model == "glm-4-flash"


def test_chatglm_init_custom(chatglm_config):
    """Test ChatGLM initialization with custom values."""
    config = chatglm_config.copy()
    config.update({
        "api_base": "https://custom.api.base",
        "model": "custom-model",
    })
    llm = ChatGLMLLM(config)
    assert llm.api_base == "https://custom.api.base"
    assert llm.model == "custom-model"


def test_chatglm_build_headers(chatglm_config):
    """Test build_headers method."""
    llm = ChatGLMLLM(chatglm_config)
    headers = llm.build_headers()
    assert headers["Content-Type"] == "application/json"
    assert headers["Authorization"] == "Bearer test-api-key"


def test_chatglm_format_messages_simple(chatglm_config, sample_message):
    """Test format_messages method with simple message."""
    llm = ChatGLMLLM(chatglm_config)
    payload = llm.format_messages(sample_message)

    assert payload["model"] == llm.model
    assert payload["max_tokens"] == chatglm_config["max_tokens"]
    assert payload["temperature"] == chatglm_config["temperature"]
    assert payload["top_p"] == chatglm_config["top_p"]
    assert len(payload["messages"]) == 1
    assert payload["messages"][0]["role"] == "user"
    assert payload["messages"][0]["content"] == sample_message


def test_chatglm_format_messages_with_history(chatglm_config, sample_message, sample_history):
    """Test format_messages method with chat history."""
    llm = ChatGLMLLM(chatglm_config)
    payload = llm.format_messages(sample_message, history=sample_history)

    assert len(payload["messages"]) == len(sample_history) + 1
    for i, msg in enumerate(sample_history):
        assert payload["messages"][i]["role"] == msg["role"]
        assert payload["messages"][i]["content"] == msg["content"]

    last_message = payload["messages"][-1]
    assert last_message["role"] == "user"
    assert last_message["content"] == sample_message


def test_chatglm_get_required_config():
    """Test get_required_config method."""
    config = ChatGLMLLM.get_required_config()
    assert isinstance(config, dict)
    assert "api_base" in config
    assert "model" in config
    assert "api_key" in config
    assert "max_tokens" in config

    # Check default values
    assert config["api_base"][0] == "https://open.bigmodel.cn/api/paas/v4"
    assert config["model"][0] == "glm-4-flash"
