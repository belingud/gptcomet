"""Test Ollama LLM."""
import pytest

from gptcomet.llms.ollama import OllamaLLM


@pytest.fixture
def ollama_config(base_config):
    """Ollama config fixture."""
    config = base_config.copy()
    config.update({
        "seed": 42,
        "top_k": 10,
        "num_gpu": 1,
        "main_gpu": 4,
    })
    return config


def test_ollama_init(ollama_config):
    """Test Ollama initialization."""
    llm = OllamaLLM(ollama_config)
    assert llm.api_base == "http://localhost:11434/api"
    assert llm.model == "llama2"
    assert llm.completion_path == "generate"
    assert llm.answer_path == "response"
    assert llm.seed == 42
    assert llm.top_k == 10
    assert llm.num_gpu == 1
    assert llm.main_gpu == 4


def test_ollama_init_custom(ollama_config):
    """Test Ollama initialization with custom values."""
    config = ollama_config.copy()
    config.update({
        "api_base": "http://custom.api:11434",
        "model": "custom-model",
        "completion_path": "custom/path",
        "answer_path": "custom.path",
    })
    llm = OllamaLLM(config)
    assert llm.api_base == "http://custom.api:11434"
    assert llm.model == "custom-model"
    assert llm.completion_path == "custom/path"
    assert llm.answer_path == "custom.path"


def test_ollama_build_headers(ollama_config):
    """Test build_headers method."""
    llm = OllamaLLM(ollama_config)
    headers = llm.build_headers()
    assert headers["Content-Type"] == "application/json"


def test_ollama_format_messages_simple(ollama_config, sample_message):
    """Test format_messages method with simple message."""
    llm = OllamaLLM(ollama_config)
    payload = llm.format_messages(sample_message)

    assert payload["model"] == llm.model
    assert payload["prompt"] == sample_message
    assert payload["options"]["num_predict"] == ollama_config["max_tokens"]
    assert payload["options"]["temperature"] == ollama_config["temperature"]
    assert payload["options"]["top_k"] == ollama_config["top_k"]
    assert payload["options"]["seed"] == ollama_config["seed"]
    assert payload["options"]["num_gpu"] == ollama_config["num_gpu"]
    assert payload["options"]["main_gpu"] == ollama_config["main_gpu"]


def test_ollama_format_messages_with_history(ollama_config, sample_message, sample_history):
    """Test format_messages method with chat history."""
    llm = OllamaLLM(ollama_config)
    payload = llm.format_messages(sample_message, history=sample_history)

    # Ollama concatenates history into a single prompt
    assert payload["prompt"].endswith(sample_message)


def test_ollama_get_required_config():
    """Test get_required_config method."""
    config = OllamaLLM.get_required_config()
    assert isinstance(config, dict)
    assert "api_base" in config
    assert "model" in config
    assert "max_tokens" in config

    # Check default values
    assert config["api_base"][0] == "http://localhost:11434/api"
    assert config["model"][0] == "llama2"
