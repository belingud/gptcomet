"""Test Azure OpenAI LLM."""
import pytest

from gptcomet.llms.azure import AzureLLM


@pytest.fixture
def azure_config(base_config):
    """Azure config fixture."""
    config = base_config.copy()
    config.update({
        "deployment_name": "test-deployment",
        "api_version": "2024-02-15-preview",
    })
    return config


def test_azure_init(azure_config):
    """Test Azure initialization."""
    llm = AzureLLM(azure_config)
    assert llm.deployment_name == "test-deployment"
    assert llm.api_version == "2024-02-15-preview"
    assert llm.model == "gpt-4o"


def test_azure_init_custom(azure_config):
    """Test Azure initialization with custom values."""
    config = azure_config.copy()
    config.update({
        "api_base": "https://custom.openai.azure.com",
        "model": "custom-model",
        "deployment_name": "custom-deployment",
        "api_version": "custom-version",
    })
    llm = AzureLLM(config)
    assert llm.api_base == "https://custom.openai.azure.com"
    assert llm.model == "custom-model"
    assert llm.deployment_name == "custom-deployment"
    assert llm.api_version == "custom-version"


def test_azure_init_missing_deployment(base_config):
    """Test Azure initialization with missing deployment name."""
    with pytest.raises(ValueError, match="deployment_name is required"):
        AzureLLM(base_config)


def test_azure_build_headers(azure_config):
    """Test build_headers method."""
    llm = AzureLLM(azure_config)
    headers = llm.build_headers()
    assert headers["Content-Type"] == "application/json"
    assert headers["api-key"] == "test-api-key"
    assert headers["api-version"] == "2024-02-15-preview"


def test_azure_format_messages_simple(azure_config, sample_message):
    """Test format_messages method with simple message."""
    llm = AzureLLM(azure_config)
    payload = llm.format_messages(sample_message)

    assert payload["messages"][0]["role"] == "user"
    assert payload["messages"][0]["content"] == sample_message
    assert payload["max_tokens"] == azure_config["max_tokens"]
    assert payload["temperature"] == azure_config["temperature"]
    assert payload["top_p"] == azure_config["top_p"]


def test_azure_format_messages_with_history(azure_config, sample_message, sample_history):
    """Test format_messages method with chat history."""
    llm = AzureLLM(azure_config)
    payload = llm.format_messages(sample_message, history=sample_history)

    assert len(payload["messages"]) == len(sample_history) + 1
    for i, msg in enumerate(sample_history):
        assert payload["messages"][i]["role"] == msg["role"]
        assert payload["messages"][i]["content"] == msg["content"]

    last_message = payload["messages"][-1]
    assert last_message["role"] == "user"
    assert last_message["content"] == sample_message


def test_azure_format_messages_with_penalties(azure_config, sample_message):
    """Test format_messages method with frequency and presence penalties."""
    config = azure_config.copy()
    config.update({
        "frequency_penalty": 0.5,
        "presence_penalty": 0.3,
    })
    
    llm = AzureLLM(config)
    payload = llm.format_messages(sample_message)

    assert payload["frequency_penalty"] == 0.5
    assert payload["presence_penalty"] == 0.3


def test_azure_build_url(azure_config):
    """Test build_url method."""
    llm = AzureLLM(azure_config)
    url = llm.build_url()
    assert "deployments/test-deployment" in url
    assert "openai/deployments" in url


def test_azure_get_required_config():
    """Test get_required_config method."""
    config = AzureLLM.get_required_config()
    assert isinstance(config, dict)
    assert "api_base" in config
    assert "model" in config
    assert "api_key" in config
    assert "deployment_name" in config
    assert "api_version" in config
    assert "max_tokens" in config

    # Check default values
    assert config["model"][0] == "gpt-4o"
    assert config["api_version"][0] == "2024-02-15-preview"
